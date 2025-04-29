package coreapi

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"fmt"
)

// SignData generates a SHA-256 signature with RSA private key
func GenerateSignature(message, privateKeyPath string) (string, error) {
	// Read private key file
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read private key: %v", err)
	}

	// Decode PEM format
	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		return "", fmt.Errorf("failed to decode private key PEM")
	}

	// Parse RSA private key
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse RSA private key: %v", err)
	}

	privateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("failed to cast parsed key to RSA private key")
	}

	// Compute SHA-256 hash
	hashed := sha256.Sum256([]byte(message))

	// Sign the hash using RSA PKCS1v15
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign data: %v", err)
	}

	// Encode to Base64
	return base64.StdEncoding.EncodeToString(signature), nil
}

func GenerateBalanceSignature(countryCode, accountId string) (string, error) {
	// Generate signature
	signature, err := GenerateSignature(countryCode+accountId, "app/keys/privatekey.pem")
	if err != nil {
		return "", fmt.Errorf("failed to generate signature: %v", err)
	}
	return signature, nil
}
