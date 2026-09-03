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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strings"
	"time"
)

const (
	// HeaderTimestamp carries the ISO8601 request timestamp that is part of the signed message.
	HeaderTimestamp = "X-YC-Timestamp"
	// HeaderWebhookSignature carries the base64 HMAC-SHA256 of a webhook body.
	HeaderWebhookSignature = "X-YC-Signature"

	authScheme      = "YcHmacV1"
	timestampLayout = "2006-01-02T15:04:05.000Z"
)

// SignatureMessage builds the message Yellow Card expects to be signed:
// timestamp + request path (without host) + upper-case method + for POST
// and PUT requests with a body, the base64 encoded SHA-256 of the body.
func SignatureMessage(timestamp, path, method string, body []byte) string {
	method = strings.ToUpper(method)
	msg := timestamp + path + method
	if len(body) > 0 && (method == http.MethodPost || method == http.MethodPut) {
		sum := sha256.Sum256(body)
		msg += base64.StdEncoding.EncodeToString(sum[:])
	}
	return msg
}

// SignRequest sets the X-YC-Timestamp and Authorization headers on req using
// the YcHmacV1 scheme: "YcHmacV1 {apiKey}:{base64(HMAC-SHA256(secret, message))}".
func SignRequest(req *http.Request, rawBody []byte, apiKey, secret string, now time.Time) {
	ts := now.UTC().Format(timestampLayout)
	sig := hmacBase64(secret, []byte(SignatureMessage(ts, req.URL.EscapedPath(), req.Method, rawBody)))
	req.Header.Set(HeaderTimestamp, ts)
	req.Header.Set("Authorization", authScheme+" "+apiKey+":"+sig)
}

// WebhookSignature computes the value Yellow Card places in X-YC-Signature:
// base64(HMAC-SHA256(secret, rawBody)).
func WebhookSignature(secret string, body []byte) string {
	return hmacBase64(secret, body)
}

// VerifyWebhookSignature reports whether header is a valid signature of body
// under secret. Empty inputs never verify.
func VerifyWebhookSignature(body []byte, header, secret string) bool {
	if header == "" || secret == "" {
		return false
	}
	expected := WebhookSignature(secret, body)
	return hmac.Equal([]byte(expected), []byte(strings.TrimSpace(header)))
}

func hmacBase64(secret string, data []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(data)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
