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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antinvestor/service-payments/apps/integrations/airtel/service/client"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlexID_StringOrObject(t *testing.T) {
	tests := []struct {
		name string
		body string
		want client.FlexID
	}{
		{name: "string", body: `{"transaction":{"id":"tx-1","airtel_money_id":"AM.1234"}}`, want: "AM.1234"},
		{name: "object", body: `{"transaction":{"id":"tx-1","airtel_money_id":{"id":"AM.5678"}}}`, want: "AM.5678"},
		{name: "number", body: `{"transaction":{"id":"tx-1","airtel_money_id":987}}`, want: "987"},
		{name: "null", body: `{"transaction":{"id":"tx-1","airtel_money_id":null}}`, want: ""},
		{name: "absent", body: `{"transaction":{"id":"tx-1"}}`, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cb client.CollectionCallbackBody
			require.NoError(t, json.Unmarshal([]byte(tt.body), &cb))
			assert.Equal(t, tt.want, cb.Transaction.AirtelMoneyID)
			assert.Equal(t, "tx-1", cb.Transaction.ID)
		})
	}
}

func TestCallbackAmount_StringOrNumber(t *testing.T) {
	var cb client.CollectionCallbackBody
	require.NoError(t, json.Unmarshal([]byte(`{"transaction":{"id":"tx-1","amount":5000}}`), &cb))
	assert.Equal(t, client.FlexString("5000"), cb.Transaction.Amount)
	require.NoError(t, json.Unmarshal([]byte(`{"transaction":{"id":"tx-1","amount":"5000.00"}}`), &cb))
	assert.Equal(t, client.FlexString("5000.00"), cb.Transaction.Amount)
}

func testKey(t *testing.T) (*rsa.PrivateKey, string, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	pemKey := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
	return key, pemKey, base64.StdEncoding.EncodeToString(der)
}

func TestEncryptPin(t *testing.T) {
	key, pemKey, b64Key := testKey(t)
	tests := []struct {
		name      string
		pin       string
		publicKey string
		wantErr   error
	}{
		{name: "PEM key", pin: "1234", publicKey: pemKey},
		{name: "base64 DER key as shown on the portal", pin: "1234", publicKey: b64Key},
		{name: "missing pin", publicKey: pemKey, wantErr: client.ErrPinNotConfigured},
		{name: "missing key", pin: "1234", wantErr: client.ErrPinNotConfigured},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.EncryptPin(tt.pin, tt.publicKey)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			cipher, err := base64.StdEncoding.DecodeString(got)
			require.NoError(t, err)
			plain, err := rsa.DecryptPKCS1v15(rand.Reader, key, cipher) //nolint:staticcheck // mirrors Airtel's decryption
			require.NoError(t, err)
			assert.Equal(t, tt.pin, string(plain))
		})
	}
	_, err := client.EncryptPin("1234", "not a key")
	require.Error(t, err)
}

type airtelServer struct {
	*httptest.Server
	disburseStatus  int
	disburseBody    string
	disbursePayload map[string]any
}

func newAirtel(t *testing.T, status int, body string) *airtelServer {
	t.Helper()
	s := &airtelServer{disburseStatus: status, disburseBody: body}
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/oauth2/token", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": 180, "token_type": "bearer"})
	})
	mux.HandleFunc("/standard/v2/disbursements/", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&s.disbursePayload)
		w.WriteHeader(s.disburseStatus)
		_, _ = w.Write([]byte(s.disburseBody))
	})
	mux.HandleFunc("/standard/v1/disbursements/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	s.Server = httptest.NewServer(mux)
	t.Cleanup(s.Close)
	return s
}

func creds(baseURL string) *client.AirtelCredentials {
	c := &client.AirtelCredentials{ClientID: "id", ClientSecret: "s", CountryCode: "UG", Currency: "UGX"}
	c.SetBaseURL(baseURL)
	return c
}

func TestDisburse_SendsEncryptedPin(t *testing.T) {
	srv := newAirtel(t, http.StatusOK, `{"data":{"transaction":{"id":"p-1","status":"TIP","reference_id":"ref"}},"status":{"success":true,"code":"200"}}`)
	cli := client.NewClientWithHTTP(srv.Client())

	resp, err := cli.Disburse(t.Context(), creds(srv.URL), &client.DisbursementRequest{
		Reference: "p-1", PhoneNumber: "770000001", Amount: "5000", Currency: "UGX", CountryCode: "UG", EncryptedPin: "enc-pin",
	})

	require.NoError(t, err)
	assert.Equal(t, "p-1", resp.Data.Transaction.ID)
	assert.Equal(t, "enc-pin", srv.disbursePayload["pin"])
	assert.Equal(t, "p-1", srv.disbursePayload["reference"])
}

func TestDisburse_RequiresPin(t *testing.T) {
	cli := client.NewClientWithHTTP(&http.Client{})
	_, err := cli.Disburse(t.Context(), creds("http://127.0.0.1:1"), &client.DisbursementRequest{Reference: "p-1"})
	require.ErrorIs(t, err, client.ErrPinNotConfigured)
}

func TestDisburse_ErrorClassification(t *testing.T) {
	tests := []struct {
		name          string
		status        int
		body          string
		wantRejected  bool
		wantUncertain bool
	}{
		{name: "success=false is rejected", status: http.StatusOK, body: `{"status":{"success":false,"code":"DP00800001001","message":"Invalid pin"}}`, wantRejected: true},
		{name: "400 is rejected", status: http.StatusBadRequest, body: `{"status":{"success":false}}`, wantRejected: true},
		{name: "503 is uncertain", status: http.StatusServiceUnavailable, body: "", wantUncertain: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newAirtel(t, tt.status, tt.body)
			cli := client.NewClientWithHTTP(srv.Client())

			_, err := cli.Disburse(t.Context(), creds(srv.URL), &client.DisbursementRequest{Reference: "p-1", EncryptedPin: "x"})

			require.Error(t, err)
			assert.Equal(t, tt.wantRejected, client.IsRejected(err))
			assert.Equal(t, tt.wantUncertain, collection.IsUncertain(err))
		})
	}
}

func TestDisbursementStatus_NotFound(t *testing.T) {
	srv := newAirtel(t, http.StatusOK, "")
	cli := client.NewClientWithHTTP(srv.Client())

	_, err := cli.DisbursementStatus(t.Context(), creds(srv.URL), "unknown")

	require.Error(t, err)
	assert.True(t, client.IsNotFound(err))
}
