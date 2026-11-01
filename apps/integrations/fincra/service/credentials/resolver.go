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

// Package credentials resolves Fincra API credentials from the settings
// service, message headers or configuration defaults. It is shared between
// the queue workers and the webhook server.
package credentials

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/v2/settings/v1/settingsv1connect"
	settingsv1 "buf.build/gen/go/antinvestor/settingz/protocolbuffers/go/settings/v1"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/config"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/client"
)

// ErrMissingCredentials indicates no Fincra secret key and business id could be resolved.
var ErrMissingCredentials = errors.New("missing required Fincra credentials (secret_key, business_id)")

// Resolver resolves Fincra credentials with a 3-level fallback:
// settings service connection -> direct headers -> config defaults.
type Resolver struct {
	settingsCli settingsv1connect.SettingsServiceClient
	cfg         *config.FincraConfig
}

// NewResolver creates a credential resolver.
func NewResolver(settingsCli settingsv1connect.SettingsServiceClient, cfg *config.FincraConfig) *Resolver {
	return &Resolver{settingsCli: settingsCli, cfg: cfg}
}

// FromHeaders resolves credentials from queue message headers, falling back
// to configuration defaults.
func (r *Resolver) FromHeaders(ctx context.Context, headers map[string]string) (*client.Credentials, error) {
	if connection := headers[config.HeaderConnectionCredentials]; connection != "" {
		return r.FromConnection(ctx, connection)
	}
	lookup := func(key, fallback string) string { return valueOrDefault(headers, key, fallback) }
	return r.validate(r.build(lookup))
}

// FromConnection resolves credentials for a named connection via the
// settings service.
func (r *Resolver) FromConnection(ctx context.Context, connection string) (*client.Credentials, error) {
	if r.settingsCli == nil {
		return nil, errors.New("settings service unavailable for connection credential lookup")
	}

	settingResp, err := r.settingsCli.Get(ctx, &settingsv1.GetRequest{
		Key: &settingsv1.Setting{
			Name:     connection,
			Object:   r.cfg.SettingsIntegrationName,
			ObjectId: r.cfg.SettingsIntegrationID,
			Module:   r.cfg.SettingsIntegrationName,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("settings lookup failed: %w", err)
	}

	var credMap map[string]string
	if err = json.Unmarshal([]byte(settingResp.GetData().GetValue()), &credMap); err != nil {
		return nil, fmt.Errorf("failed to parse settings credentials: %w", err)
	}

	// Connection values are authoritative; config only fills gaps such as
	// the environment that a connection may leave unset.
	lookup := func(key, fallback string) string { return valueOrDefault(credMap, key, fallback) }
	return r.validate(r.build(lookup))
}

// Default resolves credentials from configuration only. Used by the webhook
// server when a callback carries no connection hint.
func (r *Resolver) Default() (*client.Credentials, error) {
	lookup := func(_ string, fallback string) string { return fallback }
	return r.validate(r.build(lookup))
}

func (r *Resolver) build(lookup func(key, fallback string) string) *client.Credentials {
	return &client.Credentials{
		SecretKey:      lookup(config.HeaderSecretKey, r.cfg.SecretKey),
		PublicKey:      lookup(config.HeaderPublicKey, r.cfg.PublicKey),
		BusinessID:     lookup(config.HeaderBusinessID, r.cfg.BusinessID),
		WebhookSecret:  lookup(config.HeaderWebhookSecret, r.cfg.WebhookSecret),
		Environment:    lookup(config.HeaderEnvironment, r.cfg.Environment),
		BaseURL:        r.cfg.BaseURL,
		CollectionMode: lookup(config.HeaderCollectionMode, r.cfg.CollectionMode),
		FeeBearer:      lookup(config.HeaderFeeBearer, r.cfg.FeeBearer),
		Country:        lookup(config.HeaderCountry, r.cfg.Country),
	}
}

func (r *Resolver) validate(creds *client.Credentials) (*client.Credentials, error) {
	if creds.SecretKey == "" || creds.BusinessID == "" {
		return nil, ErrMissingCredentials
	}
	return creds, nil
}

func valueOrDefault(values map[string]string, key, fallback string) string {
	if v, ok := values[key]; ok && v != "" {
		return v
	}
	return fallback
}
