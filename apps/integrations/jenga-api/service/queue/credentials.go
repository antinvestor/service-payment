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
	"sync"
	"time"

	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	settingsv1 "buf.build/gen/go/antinvestor/settingz/protocolbuffers/go/settings/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/config"
)

const credentialCacheTTL = 5 * time.Minute

// JengaCredentials holds per-request credentials for Jenga API calls.
type JengaCredentials struct {
	MerchantCode   string
	ConsumerSecret string
	APIKey         string
	CallbackURL    string
	Environment    string
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
		return creds, nil
	}

	// Try direct headers, fall back to config
	creds := &JengaCredentials{
		MerchantCode:   headerOrDefault(headers, config.HeaderMerchantCode, r.cfg.MerchantCode),
		ConsumerSecret: headerOrDefault(headers, config.HeaderConsumerSecret, r.cfg.ConsumerSecret),
		APIKey:         headerOrDefault(headers, config.HeaderAPIKey, r.cfg.ApiKey),
		CallbackURL:    headerOrDefault(headers, config.HeaderCallbackURL, r.cfg.JengaCallbackURL),
		Environment:    headerOrDefault(headers, config.HeaderEnvironment, r.cfg.Env),
		PrivateKeyPath: headerOrDefault(headers, config.HeaderPrivateKeyPath, r.cfg.JengaPrivateKey),
	}

	if creds.MerchantCode == "" || creds.ConsumerSecret == "" || creds.APIKey == "" {
		return nil, errors.New("missing required Jenga credentials (merchant_code, consumer_secret, api_key)")
	}

	return creds, nil
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

	settingResp, err := r.settingsCli.Get(ctx, connect.NewRequest(settingReq))
	if err != nil {
		return nil, fmt.Errorf("settings lookup failed: %w", err)
	}

	var credMap map[string]string
	if err = json.Unmarshal([]byte(settingResp.Msg.GetData().GetValue()), &credMap); err != nil {
		return nil, fmt.Errorf("parse settings credentials: %w", err)
	}

	creds := &JengaCredentials{
		MerchantCode:   credMap[config.HeaderMerchantCode],
		ConsumerSecret: credMap[config.HeaderConsumerSecret],
		APIKey:         credMap[config.HeaderAPIKey],
		CallbackURL:    credMap[config.HeaderCallbackURL],
		Environment:    credMap[config.HeaderEnvironment],
		PrivateKeyPath: credMap[config.HeaderPrivateKeyPath],
	}

	return creds, nil
}

func headerOrDefault(headers map[string]string, key, fallback string) string {
	if v, ok := headers[key]; ok && v != "" {
		return v
	}
	return fallback
}
