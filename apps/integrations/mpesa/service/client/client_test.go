// Copyright 2023-2026 Ant Investor Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antinvestor/service-payments/apps/integrations/mpesa/service/client"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type darajaServer struct {
	*httptest.Server
	b2cStatus int
	b2cBody   string
	statusReq map[string]any
}

func newDaraja(t *testing.T, b2cStatus int, b2cBody string) *darajaServer {
	t.Helper()
	s := &darajaServer{b2cStatus: b2cStatus, b2cBody: b2cBody}
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/v1/generate", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": "3599"})
	})
	mux.HandleFunc("/mpesa/b2c/v1/paymentrequest", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(s.b2cStatus)
		_, _ = w.Write([]byte(s.b2cBody))
	})
	mux.HandleFunc("/mpesa/transactionstatus/v1/query", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&s.statusReq)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"OriginatorConversationID": "oc-1", "ConversationID": "AG_2", "ResponseCode": "0", "ResponseDescription": "Accept the service request successfully.",
		})
	})
	s.Server = httptest.NewServer(mux)
	t.Cleanup(s.Close)
	return s
}

func creds(baseURL string) *client.MpesaCredentials {
	c := &client.MpesaCredentials{ConsumerKey: "k", ConsumerSecret: "s", Shortcode: "600638", InitiatorName: "init", SecurityCredential: "sec"}
	c.SetBaseURL(baseURL)
	return c
}

func TestB2CPayment_ErrorClassification(t *testing.T) {
	tests := []struct {
		name          string
		status        int
		body          string
		wantErr       bool
		wantRejected  bool
		wantUncertain bool
	}{
		{name: "accepted", status: http.StatusOK, body: `{"ConversationID":"AG_1","OriginatorConversationID":"p-1","ResponseCode":"0","ResponseDescription":"ok"}`},
		{name: "200 with non-zero response code is rejected", status: http.StatusOK, body: `{"ResponseCode":"1","ResponseDescription":"no"}`, wantErr: true, wantRejected: true},
		{name: "400 is rejected", status: http.StatusBadRequest, body: `{"errorCode":"400.002.02","errorMessage":"Bad Request"}`, wantErr: true, wantRejected: true},
		{name: "500 is uncertain", status: http.StatusInternalServerError, body: `{"errorCode":"500.001.1001"}`, wantErr: true, wantUncertain: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newDaraja(t, tt.status, tt.body)
			cli := client.NewClientWithHTTP(srv.Client())

			resp, err := cli.B2CPayment(t.Context(), creds(srv.URL), &client.B2CRequest{OriginatorConversationID: "p-1", Amount: "10"})

			if !tt.wantErr {
				require.NoError(t, err)
				assert.Equal(t, "AG_1", resp.ConversationID)
				return
			}
			require.Error(t, err)
			assert.Equal(t, tt.wantRejected, client.IsRejected(err))
			assert.Equal(t, tt.wantUncertain, collection.IsUncertain(err))
		})
	}
}

func TestTransactionStatus_PostsQuery(t *testing.T) {
	srv := newDaraja(t, http.StatusOK, "")
	cli := client.NewClientWithHTTP(srv.Client())
	c := creds(srv.URL)

	resp, err := cli.TransactionStatus(t.Context(), c,
		client.NewTransactionStatusRequest(c, "LGR019G3J2", "https://hooks/result", "https://hooks/timeout"))

	require.NoError(t, err)
	assert.Equal(t, "AG_2", resp.ConversationID)
	assert.Equal(t, "TransactionStatusQuery", srv.statusReq["CommandID"])
	assert.Equal(t, "LGR019G3J2", srv.statusReq["TransactionID"])
	assert.Equal(t, "600638", srv.statusReq["PartyA"])
	assert.Equal(t, "4", srv.statusReq["IdentifierType"])
	assert.Equal(t, "init", srv.statusReq["Initiator"])
	assert.Equal(t, "https://hooks/result", srv.statusReq["ResultURL"])
}

func TestB2CPayment_ConnectionFailureIsUncertain(t *testing.T) {
	srv := newDaraja(t, http.StatusOK, "")
	url := srv.URL
	srv.Close()
	cli := client.NewClientWithHTTP(&http.Client{})

	_, err := cli.B2CPayment(t.Context(), creds(url), &client.B2CRequest{})

	require.Error(t, err)
	assert.True(t, collection.IsUncertain(err))
	assert.False(t, client.IsRejected(err))
}
