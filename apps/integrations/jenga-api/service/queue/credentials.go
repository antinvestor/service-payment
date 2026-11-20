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

package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/v2/settings/v1/settingsv1connect"
	settingsv1 "buf.build/gen/go/antinvestor/settingz/protocolbuffers/go/settings/v1"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/config"
)

const credentialCacheTTL = 5 * time.Minute

// knownJengaHosts are the hosts a Jenga base URL may point at. Anything
// else is refused before a request is signed, so a tenant setting cannot
// turn the service into a signing oracle for an arbitrary host.
var knownJengaHosts = map[string]bool{
	"uat.finserve.africa": true,
	"api.finserve.africa": true,
}

var (
	errTenantKeyRequired = errors.New(
		"a Jenga settings connection must supply its own private key as X-JENGA_PRIVATE_KEY_PEM; " +
			"a key path is never accepted from a tenant and the service key only signs for the service's own credentials")
	errBaseURLNotAllowed = errors.New("jenga base URL host is not an allowed Jenga host")
)

// JengaCredentials holds per-request credentials for Jenga API calls.
type JengaCredentials struct {
	MerchantCode   string
	ConsumerSecret string
	APIKey         string
	CallbackURL    string
	Environment    string
	BaseURL        string
	PrivateKeyPEM  string
	PrivateKeyPath string
}

type credsCacheEntry struct {
	creds  *JengaCredentials
	expiry time.Time
}

// credentialResolver provides shared credential extraction logic with caching.
type credentialResolver struct {
	settingsCli settingsv1connect.SettingsServiceClient
	cfg         *config.JengaConfig
	mu          sync.RWMutex
	cache       map[string]*credsCacheEntry
}

func (r *credentialResolver) extractCredentials(
	ctx context.Context,
	headers map[string]string,
) (*JengaCredentials, error) {
	// Try settings service lookup first (per-tenant), with cache
	connection, ok := headers[config.HeaderConnectionCredentials]
	if ok && connection != "" {
		if r.settingsCli == nil {
			return nil, errors.New("settings client not configured but connection credentials header is present")
		}
		creds, err := r.cachedCredentialsFromSettings(ctx, connection)
		if err != nil {
			return nil, err
		}
		if creds.MerchantCode == "" || creds.ConsumerSecret == "" || creds.APIKey == "" {
			return nil, errors.New(
				"incomplete Jenga credentials from settings (merchant_code, consumer_secret, api_key)",
			)
		}
		if err = r.checkBaseURL(creds.BaseURL); err != nil {
			return nil, err
		}
		return creds, nil
	}

	// Try direct headers, fall back to config
	creds := &JengaCredentials{
		MerchantCode:   headerOrDefault(headers, config.HeaderMerchantCode, r.cfg.MerchantCode),
		ConsumerSecret: headerOrDefault(headers, config.HeaderConsumerSecret, r.cfg.ConsumerSecret),
		APIKey:         headerOrDefault(headers, config.HeaderAPIKey, r.cfg.ApiKey),
		CallbackURL:    headerOrDefault(headers, config.HeaderCallbackURL, r.cfg.JengaCallbackURL),
		Environment:    headerOrDefault(headers, config.HeaderEnvironment, r.cfg.Environment),
		BaseURL:        headerOrDefault(headers, config.HeaderBaseURL, r.cfg.BaseURL),
		PrivateKeyPEM:  headerOrDefault(headers, config.HeaderPrivateKeyPEM, r.cfg.JengaPrivateKeyPEM),
		// A key path is only ever the service's own: a header could otherwise
		// name the service key file (or any readable file) as "its" key.
		PrivateKeyPath: r.cfg.JengaPrivateKey,
	}

	if creds.MerchantCode == "" || creds.ConsumerSecret == "" || creds.APIKey == "" {
		return nil, errors.New("missing required Jenga credentials (merchant_code, consumer_secret, api_key)")
	}
	// Headers that override the service's merchant, keys or host must bring
	// their own signing key; the service key signs only for the service's
	// own credentials.
	overridden := creds.MerchantCode != r.cfg.MerchantCode || creds.ConsumerSecret != r.cfg.ConsumerSecret ||
		creds.APIKey != r.cfg.ApiKey || creds.BaseURL != r.cfg.BaseURL || creds.Environment != r.cfg.Environment
	ownKey := headers[config.HeaderPrivateKeyPEM] != ""
	if overridden {
		if !ownKey {
			return nil, errTenantKeyRequired
		}
		// The service key file never backs an overridden connection.
		creds.PrivateKeyPath = ""
	}
	if err := r.checkBaseURL(creds.BaseURL); err != nil {
		return nil, err
	}

	return creds, nil
}

// checkBaseURL refuses a base URL whose host is not a known Jenga host or
// on the configured allowlist. An empty base URL resolves to a known host.
func (r *credentialResolver) checkBaseURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
		return fmt.Errorf("%w: %q", errBaseURLNotAllowed, raw)
	}
	host := strings.ToLower(u.Host)
	if knownJengaHosts[host] && u.Scheme == "https" {
		return nil
	}
	for _, allowed := range strings.Split(r.cfg.BaseURLAllowlist, ",") {
		if allowed = strings.ToLower(strings.TrimSpace(allowed)); allowed != "" && allowed == host {
			return nil
		}
	}
	return fmt.Errorf("%w: %q", errBaseURLNotAllowed, raw)
}

func (r *credentialResolver) cachedCredentialsFromSettings(
	ctx context.Context,
	connection string,
) (*JengaCredentials, error) {
	// Fast path: read from cache
	r.mu.RLock()
	if r.cache != nil {
		if entry, ok := r.cache[connection]; ok && time.Now().Before(entry.expiry) {
			creds := entry.creds
			r.mu.RUnlock()
			return creds, nil
		}
	}
	r.mu.RUnlock()

	// Slow path: fetch from settings service
	creds, err := r.credentialsFromSettings(ctx, connection)
	if err != nil {
		return nil, err
	}

	// Cache the result
	r.mu.Lock()
	if r.cache == nil {
		r.cache = make(map[string]*credsCacheEntry)
	}
	r.cache[connection] = &credsCacheEntry{
		creds:  creds,
		expiry: time.Now().Add(credentialCacheTTL),
	}
	r.mu.Unlock()

	return creds, nil
}

func (r *credentialResolver) credentialsFromSettings(
	ctx context.Context,
	connection string,
) (*JengaCredentials, error) {
	settingReq := &settingsv1.GetRequest{
		Key: &settingsv1.Setting{
			Name:     connection,
			Object:   r.cfg.SettingsIntegrationName,
			ObjectId: r.cfg.SettingsIntegrationID,
			Module:   r.cfg.SettingsIntegrationName,
		},
	}

	settingResp, err := r.settingsCli.Get(ctx, settingReq)
	if err != nil {
		return nil, fmt.Errorf("settings lookup failed: %w", err)
	}

	var credMap map[string]string
	if err = json.Unmarshal([]byte(settingResp.GetData().GetValue()), &credMap); err != nil {
		return nil, fmt.Errorf("parse settings credentials: %w", err)
	}

	creds := &JengaCredentials{
		MerchantCode:   credMap[config.HeaderMerchantCode],
		ConsumerSecret: credMap[config.HeaderConsumerSecret],
		APIKey:         credMap[config.HeaderAPIKey],
		CallbackURL:    credMap[config.HeaderCallbackURL],
		Environment:    credMap[config.HeaderEnvironment],
		BaseURL:        credMap[config.HeaderBaseURL],
		PrivateKeyPEM:  credMap[config.HeaderPrivateKeyPEM],
	}
	// The service's signing key is never lent to a tenant connection: a
	// tenant that signs must hold its own key, and only as PEM. A path from
	// settings could name the service key file on this host.
	if creds.PrivateKeyPEM == "" {
		return nil, errTenantKeyRequired
	}
	if creds.Environment == "" {
		creds.Environment = r.cfg.Environment
	}
	if creds.BaseURL == "" {
		creds.BaseURL = r.cfg.BaseURL
	}
	if creds.CallbackURL == "" {
		creds.CallbackURL = r.cfg.JengaCallbackURL
	}

	return creds, nil
}

func headerOrDefault(headers map[string]string, key, fallback string) string {
	if v, ok := headers[key]; ok && v != "" {
		return v
	}
	return fallback
}
