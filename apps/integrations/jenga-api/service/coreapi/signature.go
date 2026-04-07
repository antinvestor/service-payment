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

// testMode is a flag to skip actual signature validation during tests.
//
//nolint:gochecknoglobals // Required for test mode flag
var testMode = false

// SetTestMode sets the test mode flag for signature generation.
func SetTestMode(enabled bool) {
	testMode = enabled
}

// IsTestMode returns whether test mode is enabled.
func IsTestMode() bool {
	return testMode
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
	//nolint:gosec // G304: path is from trusted config, validated above
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
