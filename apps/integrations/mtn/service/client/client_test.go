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
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antinvestor/service-payments/apps/integrations/mtn/service/client"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mtnServer fakes the token and transfer endpoints and records headers.
type mtnServer struct {
	*httptest.Server
	transferStatus int
	headers        []http.Header
}

func newMtnServer(t *testing.T, transferStatus int) *mtnServer {
	t.Helper()
	s := &mtnServer{transferStatus: transferStatus}
	mux := http.NewServeMux()
	mux.HandleFunc("/disbursement/token/", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": 3600})
	})
	mux.HandleFunc("/disbursement/v1_0/transfer", func(w http.ResponseWriter, r *http.Request) {
		s.headers = append(s.headers, r.Header.Clone())
		w.WriteHeader(s.transferStatus)
	})
	mux.HandleFunc("/disbursement/v1_0/transfer/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	s.Server = httptest.NewServer(mux)
	t.Cleanup(s.Close)
	return s
}

func creds(baseURL string) *client.MtnCredentials {
	c := &client.MtnCredentials{SubscriptionKey: "sub", APIUser: "u", APIKey: "k", TargetEnvironment: "mtnuganda"}
	c.SetBaseURL(baseURL)
	return c
}

func TestTransfer_SendsTargetEnvironmentAndReference(t *testing.T) {
	srv := newMtnServer(t, http.StatusAccepted)
	cli := client.NewClientWithHTTP(srv.Client())

	err := cli.Transfer(t.Context(), creds(srv.URL), &client.TransferRequest{
		ReferenceID: "ref-1", Amount: "10", Currency: "UGX", Payee: client.Party{PartyIDType: "MSISDN", PartyID: "256770000001"},
	})

	require.NoError(t, err)
	require.Len(t, srv.headers, 1)
	assert.Equal(t, "mtnuganda", srv.headers[0].Get("X-Target-Environment"))
	assert.Equal(t, "ref-1", srv.headers[0].Get("X-Reference-Id"))
	assert.Equal(t, "sub", srv.headers[0].Get("Ocp-Apim-Subscription-Key"))
}

func TestTransfer_ErrorClassification(t *testing.T) {
	tests := []struct {
		name          string
		status        int
		wantUncertain bool
		wantDuplicate bool
		wantRejected  bool
	}{
		{name: "409 is a duplicate", status: http.StatusConflict, wantDuplicate: true},
		{name: "400 is rejected", status: http.StatusBadRequest, wantRejected: true},
		{name: "500 is uncertain", status: http.StatusInternalServerError, wantUncertain: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newMtnServer(t, tt.status)
			cli := client.NewClientWithHTTP(srv.Client())

			err := cli.Transfer(t.Context(), creds(srv.URL), &client.TransferRequest{ReferenceID: "ref-1"})

			require.Error(t, err)
			assert.Equal(t, tt.wantUncertain, collection.IsUncertain(err))
			assert.Equal(t, tt.wantDuplicate, client.IsDuplicate(err))
			assert.Equal(t, tt.wantRejected, client.IsRejected(err))
		})
	}
}

func TestGetTransferStatus_NotFound(t *testing.T) {
	srv := newMtnServer(t, http.StatusAccepted)
	cli := client.NewClientWithHTTP(srv.Client())

	_, err := cli.GetTransferStatus(t.Context(), creds(srv.URL), "ref-unknown")

	require.Error(t, err)
	assert.True(t, client.IsNotFound(err))
	assert.False(t, collection.IsUncertain(err))
}

func TestTransfer_ConnectionFailureIsUncertain(t *testing.T) {
	srv := newMtnServer(t, http.StatusAccepted)
	url := srv.URL
	srv.Close()
	cli := client.NewClientWithHTTP(&http.Client{})

	err := cli.Transfer(t.Context(), creds(url), &client.TransferRequest{ReferenceID: "ref-1"})

	require.Error(t, err)
	assert.True(t, collection.IsUncertain(err))
	var apiErr *client.APIError
	assert.False(t, errors.As(err, &apiErr))
}

func TestTransfer_RequiresReference(t *testing.T) {
	cli := client.NewClientWithHTTP(&http.Client{})
	require.Error(t, cli.Transfer(t.Context(), creds("http://127.0.0.1:1"), &client.TransferRequest{}))
}
