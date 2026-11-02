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
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/fwtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsRejected(t *testing.T) {
	tests := map[string]struct {
		err      error
		rejected bool
	}{
		"bad request":      {&client.HTTPError{Status: 400}, true},
		"unprocessable":    {&client.HTTPError{Status: 422}, true},
		"not found":        {&client.HTTPError{Status: 404}, true},
		"conflict":         {&client.HTTPError{Status: 409}, false},
		"duplicate in 400": {&client.HTTPError{Status: 400, Body: "Duplicate reference"}, false},
		"other 'already'":  {&client.HTTPError{Status: 400, Body: "recipient already exists"}, true},
		"rate limited":     {&client.HTTPError{Status: 429}, false},
		"timeout status":   {&client.HTTPError{Status: 408}, false},
		"server error":     {&client.HTTPError{Status: 502}, false},
		"envelope refusal": {&client.APIError{Op: "charge", Message: "declined"}, true},
		"never sent":       {fmt.Errorf("wrap: %w", client.ErrInvalidRequest), true},
		"transport":        {errors.New("execute request: dial tcp: i/o timeout"), false},
		"nil":              {nil, false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.rejected, client.IsRejected(tt.err))
		})
	}
	assert.True(t, client.IsNotFound(client.ErrNotFound))
	assert.True(t, client.IsNotFound(&client.HTTPError{Status: 404}))
	assert.True(t, client.IsDuplicate(&client.HTTPError{Status: 409}))
}

func TestIdempotencyKey(t *testing.T) {
	assert.Equal(t, "prompt-1:charge", client.IdempotencyKey("prompt-1", "charge"))
	long := client.IdempotencyKey("prompt", string(make([]byte, 100)))
	assert.Len(t, long, 64)
}

func TestARetriedWriteIsPerformedOnce(t *testing.T) {
	fw := fwtest.New(t)
	cli := client.NewClient()
	ctx := client.WithIdempotencyKey(context.Background(), "prompt-x", "charge")
	// The first attempt is performed but its answer lost; the retry with the
	// same key gets Flutterwave's original answer.
	fw.Lose("POST /orchestration/direct-charges")

	ch, err := cli.CreateOrchestratorCharge(ctx, fw.Credentials(), &client.OrchestratorChargeRequest{
		Amount: 10, Currency: "KES", Reference: "prompt-x",
		Customer:      client.CustomerInput{Email: "a@example.com"},
		PaymentMethod: client.PaymentMethodInput{Type: "card", Card: &client.CardDetails{EncryptedCardNumber: "e", Nonce: "abcdefghijkl"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "prompt-x", ch.Reference)
	assert.Equal(t, 1, fw.ChargeCount())
	assert.Equal(t, 2, fw.Count("POST /orchestration/direct-charges"))
}

func TestAWriteWithoutAKeyIsNotRetried(t *testing.T) {
	fw := fwtest.New(t)
	fw.Fail("POST /transfers", http.StatusServiceUnavailable)
	_, err := client.NewClient().CreateTransfer(context.Background(), fw.Credentials(), map[string]any{"reference": "pay-1"})
	require.Error(t, err)
	assert.Equal(t, 1, fw.Count("POST /transfers"))
}

func TestFindChargeByReference(t *testing.T) {
	fw := fwtest.New(t)
	cli := client.NewClient()
	_, err := cli.FindChargeByReference(context.Background(), fw.Credentials(), "prompt-none")
	assert.True(t, client.IsNotFound(err))

	fw.Charges["chg_9"] = &client.Charge{ID: "chg_9", Reference: "prompt-9", Status: "pending"}
	ch, err := cli.FindChargeByReference(context.Background(), fw.Credentials(), "prompt-9")
	require.NoError(t, err)
	assert.Equal(t, "chg_9", ch.ID)
}

func TestTokensAreRequestedOncePerKey(t *testing.T) {
	fw := fwtest.New(t)
	cli := client.NewClient()
	fw.Charges["chg_1"] = &client.Charge{ID: "chg_1"}
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := cli.GetCharge(context.Background(), fw.Credentials(), "chg_1")
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
	assert.Equal(t, 1, fw.Tokens)
}

func TestATokenIsNeverSharedAcrossSecrets(t *testing.T) {
	fw := fwtest.New(t)
	cli := client.NewClient()
	fw.Charges["chg_1"] = &client.Charge{ID: "chg_1"}
	good := fw.Credentials()
	_, err := cli.GetCharge(context.Background(), good, "chg_1")
	require.NoError(t, err)

	other := fw.Credentials()
	other.ClientSecret = "someone-elses-guess"
	_, err = cli.GetCharge(context.Background(), other, "chg_1")
	require.NoError(t, err)
	assert.Equal(t, 2, fw.Tokens, "the second secret earned its own token")
}

func TestRevokedKeysAreARefusalNotARetry(t *testing.T) {
	fw := fwtest.New(t)
	fw.Fail("POST /token", http.StatusUnauthorized)
	_, err := client.NewClient().GetCharge(context.Background(), fw.Credentials(), "chg_1")
	require.Error(t, err)
	assert.True(t, client.IsRejected(err), "wrong keys fail the payment instead of redelivering forever")
}

func TestAnAnswerWithoutAnIDIsUnknown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			_, _ = w.Write([]byte(`{"access_token":"t","expires_in":600}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"success","data":{}}`))
	}))
	t.Cleanup(srv.Close)
	creds := &client.Credentials{
		ClientID: "id", ClientSecret: "secret", Environment: "sandbox",
		APIBaseURL: srv.URL, OAuthTokenURL: srv.URL + "/token",
	}
	_, err := client.NewClient().CreateTransfer(context.Background(), creds, map[string]any{"reference": "pay-1"})
	require.Error(t, err)
	assert.False(t, client.IsRejected(err), "the transfer may exist: never a definite refusal")
}
