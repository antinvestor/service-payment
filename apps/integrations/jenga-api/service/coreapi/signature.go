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

package coreapi

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
)

// testMode is a flag to skip actual signature validation during tests.
//
//nolint:gochecknoglobals // Required for test mode flag
var testMode atomic.Bool

// SetTestMode sets the test mode flag for signature generation.
func SetTestMode(enabled bool) {
	testMode.Store(enabled)
}

// IsTestMode returns whether test mode is enabled.
func IsTestMode() bool {
	return testMode.Load()
}

// GenerateSignature generates a SHA-256 signature with RSA private key.
//

func GenerateSignature(message, privateKeyPath string) (string, error) {
	// For tests, return a dummy signature to avoid actual RSA key parsing
	if IsTestMode() {
		return "TEST_SIGNATURE_FOR_UNIT_TESTS", nil
	}

	// SECURITY: The privateKeyPath is set from config (env var) and must not contain path traversal.
	if privateKeyPath == "" || strings.Contains(privateKeyPath, "..") {
		return "", errors.New("invalid private key path")
	}

	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read private key: %w", err)
	}

	// Decode PEM format
	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		return "", errors.New("failed to decode private key PEM")
	}

	// Parse RSA private key
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse RSA private key: %w", err)
	}

	privateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return "", errors.New("failed to cast parsed key to RSA private key")
	}

	// Compute SHA-256 hash
	hashed := sha256.Sum256([]byte(message))

	// Sign the hash using RSA PKCS1v15
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign data: %w", err)
	}

	// Encode to Base64
	return base64.StdEncoding.EncodeToString(signature), nil
}
