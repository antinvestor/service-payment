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

	settingsv1 "buf.build/gen/go/antinvestor/settingz/protocolbuffers/go/settings/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/airtel/config"
	"github.com/antinvestor/service-payments/apps/integrations/airtel/service/client"
)

// credentialResolver provides shared credential extraction logic for queue handlers.
type credentialResolver struct {
	settingsCli settingsClient
	cfg         *config.AirtelConfig
}

// settingsClient is the interface needed for settings lookups.
type settingsClient interface {
	Get(
		ctx context.Context,
		req *connect.Request[settingsv1.GetRequest],
	) (*connect.Response[settingsv1.GetResponse], error)
}

func (cr *credentialResolver) extractCredentials(
	ctx context.Context,
	headers map[string]string,
) (*client.AirtelCredentials, error) {
	connection, ok := headers[config.HeaderConnectionCredentials]
	if ok && connection != "" {
		return cr.credentialsFromSettings(ctx, connection)
	}

	creds := &client.AirtelCredentials{
		ClientID:     headerOrDefault(headers, config.HeaderClientID, cr.cfg.ClientID),
		ClientSecret: headerOrDefault(headers, config.HeaderClientSecret, cr.cfg.ClientSecret),
		CallbackURL:  headerOrDefault(headers, config.HeaderCallbackURL, cr.cfg.CallbackURL),
		CountryCode:  headerOrDefault(headers, config.HeaderCountryCode, cr.cfg.CountryCode),
		Currency:     headerOrDefault(headers, config.HeaderCurrency, cr.cfg.Currency),
		Environment:  headerOrDefault(headers, config.HeaderEnvironment, cr.cfg.Environment),
	}

	if creds.ClientID == "" || creds.ClientSecret == "" {
		return nil, errors.New("missing required Airtel Money credentials")
	}

	return creds, nil
}

func (cr *credentialResolver) credentialsFromSettings(
	ctx context.Context,
	connection string,
) (*client.AirtelCredentials, error) {
	settingReq := &settingsv1.GetRequest{
		Key: &settingsv1.Setting{
			Name:     connection,
			Object:   cr.cfg.SettingsIntegrationName,
			ObjectId: cr.cfg.SettingsIntegrationID,
			Module:   cr.cfg.SettingsIntegrationName,
		},
	}

	settingResp, err := cr.settingsCli.Get(ctx, connect.NewRequest(settingReq))
	if err != nil {
		return nil, fmt.Errorf("settings lookup failed: %w", err)
	}

	var credMap map[string]string
	if err = json.Unmarshal([]byte(settingResp.Msg.GetData().GetValue()), &credMap); err != nil {
		return nil, fmt.Errorf("failed to parse settings credentials: %w", err)
	}

	return &client.AirtelCredentials{
		ClientID:     credMap[config.HeaderClientID],
		ClientSecret: credMap[config.HeaderClientSecret],
		CallbackURL:  credMap[config.HeaderCallbackURL],
		CountryCode:  credMap[config.HeaderCountryCode],
		Currency:     credMap[config.HeaderCurrency],
		Environment:  credMap[config.HeaderEnvironment],
	}, nil
}
