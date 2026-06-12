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

package client //nolint:testpackage // white-box tests: need access to unexported signRequest, buildSignatureBase, parsePrivateKey

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

// generateP256Key generates a fresh P-256 key for tests.
func generateP256Key(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	return key
}

// generateP384Key generates a P-384 key to test curve rejection.
func generateP384Key(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)
	return key
}

// marshalPKCS8PEM encodes key as PKCS#8 PEM.
func marshalPKCS8PEM(t *testing.T, key *ecdsa.PrivateKey) string {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

// marshalSEC1PEM encodes key as SEC1 PEM.
func marshalSEC1PEM(t *testing.T, key *ecdsa.PrivateKey) string {
	t.Helper()
	der, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)
	return string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der}))
}

// makeRequest builds an *http.Request for use in tests.
func makeRequest(t *testing.T, rawURL string, body []byte) *http.Request {
	t.Helper()
	u, err := url.Parse(rawURL)
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// ─── Content-Digest ──────────────────────────────────────────────────────────

func TestContentDigest(t *testing.T) {
	body := []byte(`{"depositId":"8917c345","amount":"15","currency":"ZMW"}`)
	sum := sha256.Sum256(body)
	expected := "sha-256=:" + base64.StdEncoding.EncodeToString(sum[:]) + ":"

	key := generateP256Key(t)
	req := makeRequest(t, "https://api.sandbox.pawapay.io/v2/deposits", body)

	err := signRequest(req, body, "TEST_KEY", key, time.Now())
	require.NoError(t, err)
	assert.Equal(t, expected, req.Header.Get("Content-Digest"))
}

// ─── Signature base golden test ───────────────────────────────────────────────

func TestBuildSignatureBase_Golden(t *testing.T) {
	// Fixed inputs so we get a deterministic base.
	method := "POST"
	authority := "api.sandbox.pawapay.io"
	path := "/v2/deposits"
	sigDate := "2024-05-02T15:36:45.058799Z"
	digest := "sha-256=:abc123=:"
	contentType := "application/json"
	keyID := "TEST_KEY"
	created := int64(1714660605)
	expires := created + 60
	sigParams := `("@method" "@authority" "@path" "signature-date" "content-digest" "content-type");alg="ecdsa-p256-sha256";keyid="TEST_KEY";created=1714660605;expires=1714660665`

	want := strings.Join([]string{
		`"@method": POST`,
		`"@authority": api.sandbox.pawapay.io`,
		`"@path": /v2/deposits`,
		`"signature-date": 2024-05-02T15:36:45.058799Z`,
		`"content-digest": sha-256=:abc123=:`,
		`"content-type": application/json`,
		`"@signature-params": ("@method" "@authority" "@path" "signature-date" "content-digest" "content-type");alg="ecdsa-p256-sha256";keyid="TEST_KEY";created=1714660605;expires=1714660665`,
	}, "\n")

	got := buildSignatureBase(method, authority, path, sigDate, digest, contentType, sigParams)
	assert.Equal(t, want, got)

	// No trailing newline
	assert.False(t, strings.HasSuffix(got, "\n"), "signature base must not have a trailing newline")

	// Sanity: all seven lines
	assert.Equal(t, 7, strings.Count(got, "\n")+1)

	_ = keyID
	_ = expires
}

// ─── Round-trip verification ─────────────────────────────────────────────────

func TestSignRequest_RoundTrip(t *testing.T) {
	body := []byte(`{"payoutId":"f4401bd2","amount":"100","currency":"ZMW"}`)
	key := generateP256Key(t)
	now := time.Unix(1714660605, 0).UTC()

	req := makeRequest(t, "https://api.sandbox.pawapay.io/v2/payouts", body)
	err := signRequest(req, body, "MY_KEY", key, now)
	require.NoError(t, err)

	// --- reconstruct the base from emitted headers ---
	digest := req.Header.Get("Content-Digest")
	sigDate := req.Header.Get("Signature-Date")
	sigInputFull := req.Header.Get("Signature-Input") // "sig-pp=(...)"

	require.NotEmpty(t, digest)
	require.NotEmpty(t, sigDate)
	require.NotEmpty(t, sigInputFull)

	// Strip "sig-pp=" prefix to get just the params part
	sigParams := strings.TrimPrefix(sigInputFull, sigLabel+"=")

	base := buildSignatureBase(
		req.Method,
		req.URL.Host,
		req.URL.Path,
		sigDate,
		digest,
		req.Header.Get("Content-Type"),
		sigParams,
	)

	// Decode the 64-byte r||s signature
	sigFull := req.Header.Get("Signature")
	require.NotEmpty(t, sigFull)
	// Strip "sig-pp=:" prefix and ":" suffix
	sigB64 := strings.TrimPrefix(sigFull, sigLabel+"=:")
	sigB64 = strings.TrimSuffix(sigB64, ":")
	rawSig, err := base64.StdEncoding.DecodeString(sigB64)
	require.NoError(t, err)
	require.Len(t, rawSig, 64, "signature must be exactly 64 bytes (r||s)")

	r := new(big.Int).SetBytes(rawSig[:32])
	s := new(big.Int).SetBytes(rawSig[32:])

	h := sha256.Sum256([]byte(base))
	assert.True(t, ecdsa.Verify(&key.PublicKey, h[:], r, s), "signature verification must pass")
}

func TestSignRequest_TamperedBodyFails(t *testing.T) {
	originalBody := []byte(`{"payoutId":"f4401bd2","amount":"100","currency":"ZMW"}`)
	tamperedBody := []byte(`{"payoutId":"f4401bd2","amount":"999","currency":"ZMW"}`)

	key := generateP256Key(t)
	now := time.Unix(1714660605, 0).UTC()

	req := makeRequest(t, "https://api.sandbox.pawapay.io/v2/payouts", originalBody)
	err := signRequest(req, originalBody, "MY_KEY", key, now)
	require.NoError(t, err)

	// Decode original signature
	sigFull := req.Header.Get("Signature")
	sigB64 := strings.TrimPrefix(sigFull, sigLabel+"=:")
	sigB64 = strings.TrimSuffix(sigB64, ":")
	rawSig, err := base64.StdEncoding.DecodeString(sigB64)
	require.NoError(t, err)
	r := new(big.Int).SetBytes(rawSig[:32])
	s := new(big.Int).SetBytes(rawSig[32:])

	// Reconstruct signature base with tampered digest
	tamperedSum := sha256.Sum256(tamperedBody)
	tamperedDigest := "sha-256=:" + base64.StdEncoding.EncodeToString(tamperedSum[:]) + ":"
	sigParams := strings.TrimPrefix(req.Header.Get("Signature-Input"), sigLabel+"=")

	tamperedBase := buildSignatureBase(
		req.Method, req.URL.Host, req.URL.Path,
		req.Header.Get("Signature-Date"),
		tamperedDigest,
		req.Header.Get("Content-Type"),
		sigParams,
	)

	// Verification against tampered base must fail
	h := sha256.Sum256([]byte(tamperedBase))
	assert.False(t, ecdsa.Verify(&key.PublicKey, h[:], r, s), "tampered body must not verify")
}

// ─── Signature-Input format ──────────────────────────────────────────────────

var sigInputPattern = regexp.MustCompile(
	`^sig-pp=\("@method" "@authority" "@path" "signature-date" "content-digest" "content-type"\);` +
		`alg="ecdsa-p256-sha256";keyid="[^"]+";created=\d+;expires=\d+$`,
)

func TestSignRequest_SignatureInputFormat(t *testing.T) {
	body := []byte(`{"test":true}`)
	key := generateP256Key(t)
	req := makeRequest(t, "https://api.sandbox.pawapay.io/v2/deposits", body)

	err := signRequest(req, body, "MY_KEY_ID", key, time.Now())
	require.NoError(t, err)

	sigInput := req.Header.Get("Signature-Input")
	assert.Regexp(t, sigInputPattern, sigInput, "Signature-Input does not match expected format")

	// created < expires
	// extract created and expires via simple string parsing to keep test simple
	assert.Contains(t, sigInput, `keyid="MY_KEY_ID"`)
	createdIdx := strings.Index(sigInput, "created=")
	expiresIdx := strings.Index(sigInput, "expires=")
	require.True(t, createdIdx > 0 && expiresIdx > 0)
	var created, expires int64
	_, errC := fmt.Sscanf(sigInput[createdIdx:], "created=%d", &created)
	_, errE := fmt.Sscanf(sigInput[expiresIdx:], "expires=%d", &expires)
	require.NoError(t, errC)
	require.NoError(t, errE)
	assert.Less(t, created, expires, "created must be < expires")
	assert.Equal(t, int64(sigDuration), expires-created)
}

// ─── parsePrivateKey ─────────────────────────────────────────────────────────

// keyDER encodes key as PKCS#8 DER for identity comparison.
func keyDER(t *testing.T, key *ecdsa.PrivateKey) []byte {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	return der
}

func TestParsePrivateKey_PKCS8(t *testing.T) {
	key := generateP256Key(t)
	pemStr := marshalPKCS8PEM(t, key)

	parsed, err := parsePrivateKey(pemStr)
	require.NoError(t, err)
	assert.Equal(t, keyDER(t, key), keyDER(t, parsed), "parsed key must match original")
}

func TestParsePrivateKey_SEC1(t *testing.T) {
	key := generateP256Key(t)
	pemStr := marshalSEC1PEM(t, key)

	parsed, err := parsePrivateKey(pemStr)
	require.NoError(t, err)
	assert.Equal(t, keyDER(t, key), keyDER(t, parsed))
}

func TestParsePrivateKey_P384Rejected(t *testing.T) {
	key := generateP384Key(t)
	pemStr := marshalPKCS8PEM(t, key)

	_, err := parsePrivateKey(pemStr)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "P-256")
}

func TestParsePrivateKey_GarbageRejected(t *testing.T) {
	_, err := parsePrivateKey("not a pem block at all")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no PEM block")
}

func TestParsePrivateKey_WrongPEMType(t *testing.T) {
	garbage := string(pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: []byte("fake"),
	}))
	_, err := parsePrivateKey(garbage)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported PEM block type")
}
