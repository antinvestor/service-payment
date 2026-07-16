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
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

const (
	nonceLen     = 12
	gcmTagBits   = 128
	nonceCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

// GenerateNonce returns a 12-character alphanumeric nonce required by Flutterwave v4.
func GenerateNonce() (string, error) {
	var b strings.Builder
	b.Grow(nonceLen)
	max := big.NewInt(int64(len(nonceCharset)))
	for range nonceLen {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("generate nonce: %w", err)
		}
		b.WriteByte(nonceCharset[n.Int64()])
	}
	return b.String(), nil
}

// EncryptAESGCM encrypts plaintext with AES-256-GCM using a base64-encoded key
// and a 12-character nonce (used as the IV bytes). Returns base64 ciphertext+tag.
// Docs: https://developer.flutterwave.com/docs/encryption
func EncryptAESGCM(plaintext, b64Key, nonce string) (string, error) {
	if plaintext == "" {
		return "", errors.New("plaintext is required")
	}
	if b64Key == "" {
		return "", errors.New("encryption key is required")
	}
	if len(nonce) != nonceLen {
		return "", fmt.Errorf("nonce must be exactly %d characters", nonceLen)
	}

	key, err := base64.StdEncoding.DecodeString(b64Key)
	if err != nil {
		return "", fmt.Errorf("decode encryption key: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCMWithTagSize(block, gcmTagBits/8)
	if err != nil {
		return "", fmt.Errorf("aes-gcm: %w", err)
	}
	if gcm.NonceSize() != nonceLen {
		// Flutterwave uses a 12-byte ASCII nonce as IV; standard GCM is also 12.
		return "", fmt.Errorf("unexpected gcm nonce size %d", gcm.NonceSize())
	}

	ct := gcm.Seal(nil, []byte(nonce), []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ct), nil
}

// EncryptCardFields encrypts PAN, expiry, and CVV with a shared nonce.
// encryptionKey is the base64 AES key from the Flutterwave dashboard.
func EncryptCardFields(
	cardNumber, expiryMonth, expiryYear, cvv, encryptionKey string,
) (*CardDetails, error) {
	nonce, err := GenerateNonce()
	if err != nil {
		return nil, err
	}
	encNum, err := EncryptAESGCM(strings.TrimSpace(cardNumber), encryptionKey, nonce)
	if err != nil {
		return nil, fmt.Errorf("encrypt card number: %w", err)
	}
	encMon, err := EncryptAESGCM(strings.TrimSpace(expiryMonth), encryptionKey, nonce)
	if err != nil {
		return nil, fmt.Errorf("encrypt expiry month: %w", err)
	}
	encYear, err := EncryptAESGCM(strings.TrimSpace(expiryYear), encryptionKey, nonce)
	if err != nil {
		return nil, fmt.Errorf("encrypt expiry year: %w", err)
	}
	encCVV, err := EncryptAESGCM(strings.TrimSpace(cvv), encryptionKey, nonce)
	if err != nil {
		return nil, fmt.Errorf("encrypt cvv: %w", err)
	}
	return &CardDetails{
		EncryptedCardNumber:  encNum,
		EncryptedExpiryMonth: encMon,
		EncryptedExpiryYear:  encYear,
		EncryptedCVV:         encCVV,
		Nonce:                nonce,
	}, nil
}

// EncryptPIN encrypts a card PIN for PUT /charges/{id} authorization.
func EncryptPIN(pin, encryptionKey string) (nonce, encrypted string, err error) {
	nonce, err = GenerateNonce()
	if err != nil {
		return "", "", err
	}
	encrypted, err = EncryptAESGCM(strings.TrimSpace(pin), encryptionKey, nonce)
	if err != nil {
		return "", "", fmt.Errorf("encrypt pin: %w", err)
	}
	return nonce, encrypted, nil
}
