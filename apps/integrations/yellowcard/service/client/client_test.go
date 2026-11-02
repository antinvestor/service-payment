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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/service/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCreds(serverURL string) *client.Credentials {
	return &client.Credentials{
		APIKey:      "KEY",
		SecretKey:   "SECRET",
		Environment: "sandbox",
		BaseURL:     serverURL + "/business",
	}
}

func TestResolveBaseURL(t *testing.T) {
	tests := []struct {
		name  string
		creds client.Credentials
		want  string
	}{
		{"sandbox by default", client.Credentials{}, "https://sandbox.api.yellowcard.io/business"},
		{"production", client.Credentials{Environment: "Production"}, "https://api.yellowcard.io/business"},
		{"override", client.Credentials{Environment: "production", BaseURL: "http://localhost:1"}, "http://localhost:1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.creds.ResolveBaseURL())
		})
	}
	assert.Equal(t, "S", (&client.Credentials{SecretKey: "S"}).ResolveWebhookSecret())
	assert.Equal(t, "W", (&client.Credentials{SecretKey: "S", WebhookSecret: "W"}).ResolveWebhookSecret())
}

const receiveResponse = `{
  "id":"7a1f2622-29b0-4e23-bb7e-bfc9883b78fa","sequenceId":"prompt-1","status":"processing",
  "channelId":"79da4d6e","country":"UG","currency":"UGX","localAmount":5000,"amount":1.3,
  "convertedAmount":5000,"rate":3800.5,"serviceFeeAmountLocal":60,"partnerFeeAmountLocal":30,
  "source":{"accountNumber":"+2561111111111","accountType":"momo","networkId":"net-mtn"},
  "recipient":{"name":"Jane","email":"jane@example.com","country":"UG"},
  "reference":"JJ8094861","expiresAt":"2023-02-02T13:50:06.298Z","createdAt":"2023-02-02T13:40:06.298Z"
}`

func TestSubmitReceive_Momo(t *testing.T) {
	var gotBody map[string]any
	var gotAuth, gotTS, gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		gotAuth, gotTS = r.Header.Get("Authorization"), r.Header.Get(client.HeaderTimestamp)
		raw, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(raw, &gotBody))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(receiveResponse))
	}))
	defer srv.Close()

	cli := client.NewClient()
	got, err := cli.SubmitReceive(t.Context(), testCreds(srv.URL), &client.ReceiveRequest{
		SequenceID:   "prompt-1",
		LocalAmount:  5000,
		Country:      "UG",
		Currency:     "UGX",
		ChannelType:  client.ChannelTypeMomo,
		Source:       client.Source{AccountType: "momo", AccountNumber: "+2561111111111", NetworkID: "net-mtn"},
		Recipient:    client.Party{Name: "Jane", Email: "jane@example.com", Country: "UG"},
		CustomerType: client.CustomerTypeRetail,
		CustomerUID:  "cust-1",
		ForceAccept:  true,
	})
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/business/receive", gotPath)
	assert.True(t, strings.HasPrefix(gotAuth, "YcHmacV1 KEY:"), gotAuth)
	assert.NotEmpty(t, gotTS)
	assert.Equal(t, "prompt-1", gotBody["sequenceId"])
	assert.InDelta(t, 5000, gotBody["localAmount"], 0)
	assert.Equal(t, true, gotBody["forceAccept"])
	assert.Equal(t, "momo", gotBody["channelType"])
	assert.Equal(t, "momo", gotBody["source"].(map[string]any)["accountType"])
	assert.Equal(t, "cust-1", gotBody["customerUID"])
	_, hasAmount := gotBody["amount"]
	assert.False(t, hasAmount, "USD amount must be omitted when localAmount is used")

	assert.Equal(t, "7a1f2622-29b0-4e23-bb7e-bfc9883b78fa", got.ID)
	assert.Equal(t, client.StatusProcessing, got.Status)
	assert.InDelta(t, 5000, got.ConvertedAmount, 0)
	assert.InDelta(t, 3800.5, got.Rate, 0)
	assert.Equal(t, "net-mtn", got.Source.NetworkID)
	assert.Nil(t, got.BankInfo)
}

func TestSubmitReceive_BankInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"r1","sequenceId":"p1","status":"processing","currency":"NGN","country":"NG",
		  "bankInfo":{"name":"PAGA","accountNumber":"01234567890","accountName":"Ken Adams","paymentLink":"https://pay.example/x","extraKey":"v"},
		  "reference":"REF1","convertedAmount":9757.8}`))
	}))
	defer srv.Close()

	got, err := client.NewClient().SubmitReceive(t.Context(), testCreds(srv.URL), &client.ReceiveRequest{SequenceID: "p1"})
	require.NoError(t, err)
	require.NotNil(t, got.BankInfo)
	assert.Equal(t, "PAGA", got.BankInfo.Name)
	assert.Equal(t, "01234567890", got.BankInfo.AccountNumber)
	assert.Equal(t, "Ken Adams", got.BankInfo.AccountName)
	assert.Equal(t, "https://pay.example/x", got.BankInfo.PaymentLink)
	assert.Equal(t, "v", got.BankInfo.Extra["extraKey"])
	assert.Equal(t, "REF1", got.Reference)
}

func TestSubmitReceive_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":"PaymentValidationError","message":"amount must be between 0 and 1000"}`))
	}))
	defer srv.Close()

	_, err := client.NewClient().SubmitReceive(t.Context(), testCreds(srv.URL), &client.ReceiveRequest{SequenceID: "p1"})
	var apiErr *client.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusBadRequest, apiErr.HTTPStatus)
	assert.Equal(t, client.ErrCodePaymentValidation, apiErr.Code)
	assert.Equal(t, "amount must be between 0 and 1000", apiErr.Message)
	assert.False(t, client.IsNotFound(err))
	assert.False(t, client.IsDuplicate(err))
}

func TestSubmitReceive_DuplicateSequence(t *testing.T) {
	err := &client.APIError{HTTPStatus: 400, Code: "PaymentValidationError", Message: "Duplicate sequenceId"}
	assert.True(t, client.IsDuplicate(err))
	assert.False(t, client.IsDuplicate(errors.New("other")))
}

func TestGetReceiveBySequenceID_NotFound(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":"CollectionNotFoundError","message":""}`))
	}))
	defer srv.Close()

	_, err := client.NewClient().GetReceiveBySequenceID(t.Context(), testCreds(srv.URL), "seq/1")
	require.Error(t, err)
	assert.True(t, client.IsNotFound(err))
	assert.Equal(t, "/business/receive/sequence-id/seq%2F1", gotPath)
}

func TestGetChannels_RetriesOn503(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "UG", r.URL.Query().Get("country"))
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"code":"InternalServerError","message":"something went wrong"}`))
			return
		}
		_, _ = w.Write([]byte(`[{"id":"c1","country":"UG","currency":"UGX","status":"active","apiStatus":"active","channelType":"momo","rampType":"deposit","min":100,"max":5000000}]`))
	}))
	defer srv.Close()

	channels, err := client.NewClient().GetChannels(t.Context(), testCreds(srv.URL), "UG")
	require.NoError(t, err)
	require.Len(t, channels, 1)
	assert.Equal(t, "momo", channels[0].ChannelType)
	assert.InDelta(t, 5000000, channels[0].Max, 0)
	assert.Equal(t, int32(2), calls.Load())
}

func TestSubmitReceive_NoRetryOn503(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"code":"InternalServerError","message":"something went wrong"}`))
	}))
	defer srv.Close()

	_, err := client.NewClient().SubmitReceive(t.Context(), testCreds(srv.URL), &client.ReceiveRequest{SequenceID: "p1"})
	var apiErr *client.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusServiceUnavailable, apiErr.HTTPStatus)
	assert.Equal(t, int32(1), calls.Load())
}

func TestGetRatesAndNetworks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/business/rates":
			assert.Equal(t, "UGX", r.URL.Query().Get("currency"))
			_, _ = w.Write([]byte(`{"rates":[{"buy":3800.5,"sell":3700.1,"locale":"UG","rateId":"shilling","code":"UGX"}]}`))
		case "/business/networks":
			_, _ = w.Write([]byte(`[{"id":"n1","code":"MTN","name":"MTN Uganda","country":"UG","status":"active","accountNumberType":"momo","channelIds":["c1"]}]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	cli := client.NewClient()
	rates, err := cli.GetRates(t.Context(), testCreds(srv.URL), "UGX")
	require.NoError(t, err)
	require.Len(t, rates, 1)
	assert.InDelta(t, 3800.5, rates[0].Buy, 0)

	networks, err := cli.GetNetworks(t.Context(), testCreds(srv.URL), "UG")
	require.NoError(t, err)
	require.Len(t, networks, 1)
	assert.Equal(t, "MTN Uganda", networks[0].Name)
	assert.Equal(t, []string{"c1"}, networks[0].ChannelIDs)
}

func TestSubmitSendAndActions(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		_, _ = w.Write([]byte(`{"id":"s1","sequenceId":"pay-1","status":"created","currency":"GHS","country":"GH","convertedAmount":20000,"rate":14,
		  "destination":{"accountName":"Json Bourne","accountNumber":"1111111111","accountType":"bank","networkId":"n1"}}`))
	}))
	defer srv.Close()

	cli := client.NewClient()
	creds := testCreds(srv.URL)
	got, err := cli.SubmitSend(t.Context(), creds, &client.SendRequest{SequenceID: "pay-1", Reason: "other", ForceAccept: true})
	require.NoError(t, err)
	assert.Equal(t, "s1", got.ID)
	assert.Equal(t, "n1", got.Destination.NetworkID)

	_, err = cli.AcceptSend(t.Context(), creds, "s1")
	require.NoError(t, err)
	_, err = cli.GetSendBySequenceID(t.Context(), creds, "pay-1")
	require.NoError(t, err)
	_, err = cli.AcceptReceive(t.Context(), creds, "r1")
	require.NoError(t, err)
	_, err = cli.RefundReceive(t.Context(), creds, "r1")
	require.NoError(t, err)
	_, err = cli.GetReceive(t.Context(), creds, "r1")
	require.NoError(t, err)

	assert.Equal(t, []string{
		"POST /business/send",
		"POST /business/send/s1/accept",
		"GET /business/send/sequence-id/pay-1",
		"POST /business/receive/r1/accept",
		"POST /business/receive/r1/refund",
		"GET /business/receive/r1",
	}, paths)
}
