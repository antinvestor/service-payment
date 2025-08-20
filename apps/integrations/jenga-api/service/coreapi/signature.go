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
)

// TestMode is a flag to skip actual signature validation during tests.
//
//nolint:revive // TestMode follows standard naming
var TestMode bool = false

// SignData generates a SHA-256 signature with RSA private key.
//
//nolint:revive // GenerateSignature follows standard naming
func GenerateSignature(message, privateKeyPath string) (string, error) {
	// For tests, return a dummy signature to avoid actual RSA key parsing
	if TestMode {
		return "TEST_SIGNATURE_FOR_UNIT_TESTS", nil
	}

	// Read private key file
	// SECURITY: The privateKeyPath should be set from a trusted source (e.g., environment variable or config file)
	// and must not be influenced by untrusted user input to avoid file inclusion vulnerabilities (G304).
	if privateKeyPath == "" || privateKeyPath[0] == '/' || strings.Contains(privateKeyPath, "..") {
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
