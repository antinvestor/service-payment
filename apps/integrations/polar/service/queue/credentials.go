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

	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/v2/settings/v1/settingsv1connect"
	settingsv1 "buf.build/gen/go/antinvestor/settingz/protocolbuffers/go/settings/v1"
	"github.com/antinvestor/service-payments/apps/integrations/polar/config"
	"github.com/antinvestor/service-payments/apps/integrations/polar/service/client"
)

// credentialResolver picks the Polar credentials for a queue message:
// a settings connection named in the headers, else header overrides, else
// the service defaults.
type credentialResolver struct {
	settingsCli settingsv1connect.SettingsServiceClient
	cfg         *config.PolarConfig
}

func (r credentialResolver) resolve(ctx context.Context, headers map[string]string) (*client.PolarCredentials, error) {
	if connection := headers[config.HeaderConnectionCredentials]; connection != "" {
		return r.fromSettings(ctx, connection)
	}
	creds := &client.PolarCredentials{
		APIKey:         headerOrDefault(headers, config.HeaderAPIKey, r.cfg.APIKey),
		WebhookSecret:  headerOrDefault(headers, config.HeaderWebhookSecret, r.cfg.WebhookSecret),
		OrganizationID: headerOrDefault(headers, config.HeaderOrganizationID, r.cfg.OrganizationID),
		Environment:    headerOrDefault(headers, config.HeaderEnvironment, r.cfg.Environment),
	}
	if creds.APIKey == "" {
		return nil, errors.New("missing required Polar API key")
	}
	return creds, nil
}

func (r credentialResolver) fromSettings(ctx context.Context, connection string) (*client.PolarCredentials, error) {
	if r.settingsCli == nil {
		return nil, errors.New("settings client not configured; cannot resolve connection credentials")
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
	creds := &client.PolarCredentials{
		APIKey:         credMap[config.HeaderAPIKey],
		WebhookSecret:  credMap[config.HeaderWebhookSecret],
		OrganizationID: credMap[config.HeaderOrganizationID],
		Environment:    credMap[config.HeaderEnvironment],
	}
	if creds.Environment == "" {
		creds.Environment = r.cfg.Environment
	}
	if creds.APIKey == "" {
		return nil, errors.New("settings connection has no Polar API key")
	}
	return creds, nil
}

func headerOrDefault(headers map[string]string, key, fallback string) string {
	if v, ok := headers[key]; ok && v != "" {
		return v
	}
	return fallback
}
