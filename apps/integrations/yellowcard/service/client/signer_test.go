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
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/service/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignatureMessage(t *testing.T) {
	body := []byte(`{"a":1}`)
	sum := sha256.Sum256(body)
	want := "2022-01-11T15:48:37.424Z/business/receivePOST" + base64.StdEncoding.EncodeToString(sum[:])

	assert.Equal(t, want, client.SignatureMessage("2022-01-11T15:48:37.424Z", "/business/receive", "post", body))
	assert.Equal(t, "2022-01-11T15:48:37.424Z/business/channelsGET",
		client.SignatureMessage("2022-01-11T15:48:37.424Z", "/business/channels", http.MethodGet, nil))
	// A GET never hashes a body even if one is supplied.
	assert.Equal(t, "2022-01-11T15:48:37.424Z/business/channelsGET",
		client.SignatureMessage("2022-01-11T15:48:37.424Z", "/business/channels", http.MethodGet, body))
}

func TestSignRequest(t *testing.T) {
	now := time.Date(2022, 1, 11, 15, 48, 37, 424_000_000, time.UTC)
	body := []byte(`{"a":1}`)
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost,
		"https://sandbox.api.yellowcard.io/business/receive?x=1", bytes.NewReader(body))
	require.NoError(t, err)

	client.SignRequest(req, body, "key", "secret", now)

	assert.Equal(t, "2022-01-11T15:48:37.424Z", req.Header.Get(client.HeaderTimestamp))

	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write([]byte(client.SignatureMessage("2022-01-11T15:48:37.424Z", "/business/receive", "POST", body)))
	want := "YcHmacV1 key:" + base64.StdEncoding.EncodeToString(mac.Sum(nil))
	assert.Equal(t, want, req.Header.Get("Authorization"))
}

func TestWebhookSignature(t *testing.T) {
	body := []byte(`{"id":"1"}`)
	sig := client.WebhookSignature("secret", body)

	assert.True(t, client.VerifyWebhookSignature(body, sig, "secret"))
	assert.True(t, client.VerifyWebhookSignature(body, " "+sig+"\n", "secret"))
	assert.False(t, client.VerifyWebhookSignature([]byte(`{"id":"2"}`), sig, "secret"))
	assert.False(t, client.VerifyWebhookSignature(body, sig, "other"))
	assert.False(t, client.VerifyWebhookSignature(body, "", "secret"))
	assert.False(t, client.VerifyWebhookSignature(body, sig, ""))
}
