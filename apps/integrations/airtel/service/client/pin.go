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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

// ErrPinNotConfigured is returned when a disbursement is attempted without
// the wallet PIN or Airtel's public key.
var ErrPinNotConfigured = errors.New("airtel disbursement PIN or public key not configured")

// EncryptPin RSA-encrypts the disbursement PIN with Airtel's public key
// (PKCS#1 v1.5, as Airtel's v2 disbursement API expects) and returns the
// base64 ciphertext for the payload's "pin" field. The key may be PEM or
// the bare base64 DER (SubjectPublicKeyInfo or PKCS#1) Airtel's portal shows.
func EncryptPin(pin, publicKey string) (string, error) {
	pin = strings.TrimSpace(pin)
	if pin == "" || strings.TrimSpace(publicKey) == "" {
		return "", ErrPinNotConfigured
	}
	key, err := ParsePublicKey(publicKey)
	if err != nil {
		return "", err
	}
	// Airtel's disbursement API decrypts the PIN with RSA/ECB/PKCS1Padding;
	// OAEP is not accepted, so the deprecated padding is the one required.
	cipher, err := rsa.EncryptPKCS1v15(rand.Reader, key, []byte(pin)) //nolint:staticcheck // provider mandates PKCS#1 v1.5
	if err != nil {
		return "", fmt.Errorf("encrypt airtel pin: %w", err)
	}
	return base64.StdEncoding.EncodeToString(cipher), nil
}

// ParsePublicKey reads an RSA public key from PEM or base64 DER.
func ParsePublicKey(raw string) (*rsa.PublicKey, error) {
	raw = strings.TrimSpace(raw)
	var der []byte
	if block, _ := pem.Decode([]byte(raw)); block != nil {
		der = block.Bytes
	} else {
		decoded, err := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(raw), ""))
		if err != nil {
			return nil, fmt.Errorf("airtel public key is neither PEM nor base64 DER: %w", err)
		}
		der = decoded
	}
	if key, err := x509.ParsePKIXPublicKey(der); err == nil {
		rsaKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("airtel public key is not an RSA key")
		}
		return rsaKey, nil
	}
	key, err := x509.ParsePKCS1PublicKey(der)
	if err != nil {
		return nil, fmt.Errorf("parse airtel public key: %w", err)
	}
	return key, nil
}
