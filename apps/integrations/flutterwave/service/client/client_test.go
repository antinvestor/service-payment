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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyWebhookSignature_HMAC(t *testing.T) {
	c := client.NewClient()
	body := []byte(`{"type":"charge.completed","data":{"id":"chg_1"}}`)
	secret := "my-secret-hash"

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	assert.True(t, c.VerifyWebhookSignature(body, sig, secret))
	assert.False(t, c.VerifyWebhookSignature(body, "bad", secret))
	assert.False(t, c.VerifyWebhookSignature(body, sig, "other"))
}

func TestCreateOrchestratorCharge_RequiresOAuthCreds(t *testing.T) {
	cli := client.NewClient()
	_, err := cli.CreateOrchestratorCharge(context.Background(), &client.Credentials{}, &client.OrchestratorChargeRequest{
		Amount:    10,
		Currency:  "KES",
		Reference: "prompt-test1",
		Customer:  client.CustomerInput{Email: "a@b.com"},
		PaymentMethod: client.PaymentMethodInput{
			Type: "bank_transfer",
			BankTransfer: &client.BankTransferDetails{
				AccountType: "dynamic",
			},
		},
	})
	require.Error(t, err)
	// Empty creds cannot OAuth or multipay.
	assert.Error(t, err)
}

func TestCreateOrchestratorCharge_CardWithoutEncryptionExplainsEmbedded(t *testing.T) {
	cli := client.NewClient()
	_, err := cli.CreateOrchestratorCharge(context.Background(), &client.Credentials{
		ClientID:     "oauth-client-id",
		ClientSecret: "oauth-client-secret",
		Environment:  "sandbox",
	}, &client.OrchestratorChargeRequest{
		Amount:    10,
		Currency:  "KES",
		Reference: "prompt-card1",
		Customer:  client.CustomerInput{Email: "a@b.com"},
		PaymentMethod: client.PaymentMethodInput{
			Type: "card", // no encrypted Card payload
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "encrypted card")
	assert.NotContains(t, err.Error(), "FLWSECK_*")
	assert.NotContains(t, strings.ToLower(err.Error()), "set flutterwave_secret_key")
}

func TestExtractRedirectURL(t *testing.T) {
	ch := &client.Charge{
		NextAction: map[string]any{
			"type": "redirect_url",
			"redirect_url": map[string]any{
				"url": "https://example.com/pay",
			},
		},
	}
	assert.Equal(t, "https://example.com/pay", client.ExtractRedirectURL(ch))
	assert.Empty(t, client.ExtractRedirectURL(&client.Charge{}))
}
