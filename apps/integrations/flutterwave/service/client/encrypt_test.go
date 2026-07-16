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
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"testing"
)

func TestGenerateNonce(t *testing.T) {
	t.Parallel()
	n, err := GenerateNonce()
	if err != nil {
		t.Fatal(err)
	}
	if len(n) != 12 {
		t.Fatalf("nonce len=%d want 12", len(n))
	}
}

func TestEncryptAESGCMRoundTrip(t *testing.T) {
	t.Parallel()
	// 32-byte key for AES-256
	rawKey := make([]byte, 32)
	for i := range rawKey {
		rawKey[i] = byte(i + 1)
	}
	b64Key := base64.StdEncoding.EncodeToString(rawKey)
	nonce := "AbCdEfGhIjKl"
	plain := "5531886652142950"

	ct, err := EncryptAESGCM(plain, b64Key, nonce)
	if err != nil {
		t.Fatal(err)
	}
	if ct == "" || ct == plain {
		t.Fatal("expected non-empty ciphertext different from plain")
	}

	// Decrypt to verify.
	block, err := aes.NewCipher(rawKey)
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	rawCT, err := base64.StdEncoding.DecodeString(ct)
	if err != nil {
		t.Fatal(err)
	}
	pt, err := gcm.Open(nil, []byte(nonce), rawCT, nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(pt) != plain {
		t.Fatalf("got %q want %q", pt, plain)
	}
}

func TestEncryptCardFields(t *testing.T) {
	t.Parallel()
	rawKey := make([]byte, 32)
	for i := range rawKey {
		rawKey[i] = byte(i + 3)
	}
	b64Key := base64.StdEncoding.EncodeToString(rawKey)
	card, err := EncryptCardFields("5531886652142950", "09", "32", "564", b64Key)
	if err != nil {
		t.Fatal(err)
	}
	if card.Nonce == "" || card.EncryptedCardNumber == "" || card.EncryptedCVV == "" {
		t.Fatalf("incomplete card: %+v", card)
	}
}

func TestIsRetryable(t *testing.T) {
	t.Parallel()
	if !isRetryable(errString("flutterwave http 503: unavailable")) {
		t.Fatal("503 should retry")
	}
	if isRetryable(errString("flutterwave http 422: encryption")) {
		t.Fatal("422 should not retry")
	}
}

type errString string

func (e errString) Error() string { return string(e) }
