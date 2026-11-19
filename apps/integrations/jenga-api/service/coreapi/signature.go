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
	"sync"
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

// ErrNoSigningKey is returned when credentials carry neither a PEM nor a path.
var ErrNoSigningKey = errors.New("jenga signing key is not configured (private key PEM or path)")

// signer parses and caches RSA private keys keyed by their source.
type signer struct {
	mu   sync.RWMutex
	keys map[string]*rsa.PrivateKey
}

func newSigner() *signer {
	return &signer{keys: map[string]*rsa.PrivateKey{}}
}

// keyFor returns the parsed RSA key for the credentials: the PEM when
// given, else the file at PrivateKeyPath.
func (s *signer) keyFor(creds *Credentials) (*rsa.PrivateKey, error) {
	pemData := strings.TrimSpace(creds.PrivateKeyPEM)
	path := strings.TrimSpace(creds.PrivateKeyPath)
	cacheKey := pemData
	if cacheKey == "" {
		cacheKey = "path:" + path
	}
	s.mu.RLock()
	key, ok := s.keys[cacheKey]
	s.mu.RUnlock()
	if ok {
		return key, nil
	}

	var raw []byte
	switch {
	case pemData != "":
		raw = []byte(pemData)
	case path != "":
		// SECURITY: the path comes from config or settings and must not traverse.
		if strings.Contains(path, "..") {
			return nil, errors.New("invalid private key path")
		}
		var err error
		raw, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key: %w", err)
		}
	default:
		return nil, ErrNoSigningKey
	}

	key, err := parseRSAPrivateKey(raw)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.keys[cacheKey] = key
	s.mu.Unlock()
	return key, nil
}

// sign produces Jenga's base64 RSA-SHA256 signature of the concatenated fields.
func (s *signer) sign(creds *Credentials, fields ...string) (string, error) {
	if IsTestMode() {
		return "TEST_SIGNATURE_FOR_UNIT_TESTS", nil
	}
	key, err := s.keyFor(creds)
	if err != nil {
		return "", err
	}
	return signMessage(key, strings.Join(fields, ""))
}

func parseRSAPrivateKey(raw []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("failed to decode private key PEM")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		privateKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("failed to cast parsed key to RSA private key")
		}
		return privateKey, nil
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
	}
	return privateKey, nil
}

func signMessage(key *rsa.PrivateKey, message string) (string, error) {
	hashed := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign data: %w", err)
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// GenerateSignature generates a SHA-256 signature with the RSA private key
// read from privateKeyPath.
func GenerateSignature(message, privateKeyPath string) (string, error) {
	if IsTestMode() {
		return "TEST_SIGNATURE_FOR_UNIT_TESTS", nil
	}
	if privateKeyPath == "" {
		return "", ErrNoSigningKey
	}
	key, err := newSigner().keyFor(&Credentials{PrivateKeyPath: privateKeyPath})
	if err != nil {
		return "", err
	}
	return signMessage(key, message)
}
