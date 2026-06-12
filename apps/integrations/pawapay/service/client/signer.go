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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
)

const (
	sigLabel = "sig-pp"
	sigAlg   = "ecdsa-p256-sha256"

	// sigDuration is how long a signed request is valid.
	sigDuration = 60

	// ecdsaPartLen is the fixed byte length of each ECDSA component (r or s)
	// for P-256 (256 bits / 8 = 32 bytes).
	ecdsaPartLen = 32

	// ecdsaSigLen is the total length of the concatenated r||s signature bytes.
	ecdsaSigLen = ecdsaPartLen * 2
)

// signRequest adds RFC-9421 message-signature headers (Content-Digest,
// Signature-Date, Signature-Input, Signature) so pawaPay can verify the
// integrity and origin of the data we send. Only called when signing
// credentials are configured.
func signRequest(req *http.Request, body []byte, keyID string, key *ecdsa.PrivateKey, now time.Time) error {
	// Content-Digest: sha-256=:<base64(sha256(body))>:
	sum := sha256.Sum256(body)
	digest := "sha-256=:" + base64.StdEncoding.EncodeToString(sum[:]) + ":"

	// Signature-Date: RFC3339 UTC
	sigDate := now.UTC().Format(time.RFC3339Nano)

	created := now.Unix()
	expires := created + sigDuration

	authority := req.URL.Host
	path := req.URL.Path

	// Signature-Input structured field
	sigParams := fmt.Sprintf(
		`("%s" "%s" "%s" "signature-date" "content-digest" "content-type");alg="%s";keyid="%s";created=%d;expires=%d`,
		"@method", "@authority", "@path",
		sigAlg, keyID, created, expires,
	)
	sigInput := sigLabel + "=" + sigParams

	// Build signature base per RFC 9421 §2.5
	base := buildSignatureBase(req.Method, authority, path, sigDate, digest, req.Header.Get("Content-Type"), sigParams)

	// Sign: ECDSA-P256 over SHA-256(base) → 64-byte r||s
	h := sha256.Sum256([]byte(base))
	r, s, err := ecdsa.Sign(rand.Reader, key, h[:])
	if err != nil {
		return fmt.Errorf("ecdsa sign: %w", err)
	}

	// Encode as fixed-width 64-byte r||s (RFC 9421 requirement)
	rBytes := zeroPad(r, ecdsaPartLen)
	rawSig := make([]byte, ecdsaSigLen)
	copy(rawSig[:ecdsaPartLen], rBytes)
	copy(rawSig[ecdsaPartLen:], zeroPad(s, ecdsaPartLen))
	sigValue := sigLabel + "=:" + base64.StdEncoding.EncodeToString(rawSig) + ":"

	// Set headers
	req.Header.Set("Content-Digest", digest)
	req.Header.Set("Signature-Date", sigDate)
	req.Header.Set("Signature-Input", sigInput)
	req.Header.Set("Signature", sigValue)

	return nil
}

// buildSignatureBase constructs the RFC 9421 §2.5 signature base string.
// Lines are joined with \n; there is NO trailing newline.
func buildSignatureBase(method, authority, path, sigDate, digest, contentType, sigParams string) string {
	lines := []string{
		`"@method": ` + method,
		`"@authority": ` + authority,
		`"@path": ` + path,
		`"signature-date": ` + sigDate,
		`"content-digest": ` + digest,
		`"content-type": ` + contentType,
		`"@signature-params": ` + sigParams,
	}
	return strings.Join(lines, "\n")
}

// zeroPad returns the big.Int's bytes left-padded with zeros to length n.
func zeroPad(n *big.Int, length int) []byte {
	b := n.Bytes()
	if len(b) >= length {
		return b[len(b)-length:]
	}
	padded := make([]byte, length)
	copy(padded[length-len(b):], b)
	return padded
}

// parsePrivateKey parses a PEM-encoded ECDSA private key.
// Accepted PEM types: "PRIVATE KEY" (PKCS#8) and "EC PRIVATE KEY" (SEC1).
// Returns an error if the key is not on the P-256 curve.
func parsePrivateKey(pemData string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("pawapay signer: no PEM block found in private key")
	}

	var ecKey *ecdsa.PrivateKey

	switch block.Type {
	case "PRIVATE KEY": // PKCS#8
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("pawapay signer: parse PKCS#8 key: %w", err)
		}
		var ok bool
		ecKey, ok = key.(*ecdsa.PrivateKey)
		if !ok {
			return nil, errors.New("pawapay signer: PKCS#8 key is not an ECDSA key")
		}
	case "EC PRIVATE KEY": // SEC1
		var err error
		ecKey, err = x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("pawapay signer: parse SEC1 EC key: %w", err)
		}
	default:
		return nil, fmt.Errorf("pawapay signer: unsupported PEM block type %q", block.Type)
	}

	if ecKey.Curve != elliptic.P256() {
		return nil, fmt.Errorf("pawapay signer: key curve %s is not P-256", ecKey.Curve.Params().Name)
	}

	return ecKey, nil
}
