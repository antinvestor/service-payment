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

package client

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"strings"
)

// HeaderWebhookSignature carries Fincra's webhook signature.
const HeaderWebhookSignature = "signature"

// WebhookSignature is hex(HMAC-SHA512(secret, message)), Fincra's scheme.
func WebhookSignature(secret string, message []byte) string {
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write(message)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyWebhookSignature reports whether signature is Fincra's signature of
// the webhook body. Fincra signs JSON.stringify(payload), so besides the raw
// body the compact form is accepted, and so is the compact {event, data}
// object its docs define as the signed payload. An empty secret or
// signature never verifies.
func VerifyWebhookSignature(body []byte, signature, secret string) bool {
	signature = strings.ToLower(strings.TrimSpace(signature))
	if secret == "" || signature == "" {
		return false
	}
	got, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	for _, message := range signedForms(body) {
		want, _ := hex.DecodeString(WebhookSignature(secret, message))
		if hmac.Equal(got, want) {
			return true
		}
	}
	return false
}

func signedForms(body []byte) [][]byte {
	forms := [][]byte{body}
	var compact bytes.Buffer
	if json.Compact(&compact, body) == nil && !bytes.Equal(compact.Bytes(), body) {
		forms = append(forms, compact.Bytes())
	}
	var parts struct {
		Event json.RawMessage `json:"event"`
		Data  json.RawMessage `json:"data"`
	}
	if json.Unmarshal(body, &parts) == nil && len(parts.Event) > 0 && len(parts.Data) > 0 {
		var event, data bytes.Buffer
		if json.Compact(&event, parts.Event) == nil && json.Compact(&data, parts.Data) == nil {
			eventData := []byte(`{"event":` + event.String() + `,"data":` + data.String() + `}`)
			if !bytes.Equal(eventData, compact.Bytes()) {
				forms = append(forms, eventData)
			}
		}
	}
	return forms
}
