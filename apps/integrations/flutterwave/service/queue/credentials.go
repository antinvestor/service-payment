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
	"strings"

	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	settingsv1 "buf.build/gen/go/antinvestor/settingz/protocolbuffers/go/settings/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/config"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/client"
)

type credentialResolver struct {
	settingsCli settingsv1connect.SettingsServiceClient
	cfg         *config.FlutterwaveConfig
}

func (r *credentialResolver) extractCredentials(
	ctx context.Context,
	headers map[string]string,
) (*client.Credentials, error) {
	if connection, ok := headers[config.HeaderConnectionCredentials]; ok && connection != "" {
		return r.credentialsFromSettings(ctx, connection)
	}

	creds := &client.Credentials{
		ClientID:      headerOrDefault(headers, config.HeaderClientID, r.cfg.ClientID),
		ClientSecret:  headerOrDefault(headers, config.HeaderClientSecret, r.cfg.ClientSecret),
		PublicKey:     firstNonEmpty(r.cfg.PublicKey, r.cfg.ClientID),
		SecretKey:     firstNonEmpty(r.cfg.SecretKey, r.cfg.ClientSecret),
		EncryptionKey: r.cfg.EncryptionKey,
		WebhookSecret: headerOrDefault(headers, config.HeaderWebhookSecret, r.cfg.WebhookSecret),
		Environment:   headerOrDefault(headers, config.HeaderEnvironment, r.cfg.Environment),
		OAuthTokenURL: r.cfg.OAuthTokenURL,
	}
	// Accept dashboard v3 keys mapped into client_id/client_secret fields.
	if strings.HasPrefix(creds.ClientID, "FLWPUBK_") {
		creds.PublicKey = creds.ClientID
	}
	if strings.HasPrefix(creds.ClientSecret, "FLWSECK_") {
		creds.SecretKey = creds.ClientSecret
	}
	if strings.EqualFold(creds.Environment, "production") {
		creds.APIBaseURL = r.cfg.ProductionAPIBaseURL
	} else {
		creds.APIBaseURL = r.cfg.SandboxAPIBaseURL
	}
	if client.IsV3Credentials(creds) {
		if creds.SecretKey == "" {
			return nil, errors.New("missing Flutterwave secret key (FLWSECK_* via FLUTTERWAVE_CLIENT_SECRET or FLUTTERWAVE_SECRET_KEY)")
		}
		return creds, nil
	}
	if creds.ClientID == "" || creds.ClientSecret == "" {
		return nil, errors.New("missing Flutterwave v4 client_id/client_secret (FLUTTERWAVE_CLIENT_ID / FLUTTERWAVE_CLIENT_SECRET)")
	}
	return creds, nil
}

func (r *credentialResolver) credentialsFromSettings(
	ctx context.Context,
	connection string,
) (*client.Credentials, error) {
	if r.settingsCli == nil {
		return nil, errors.New("settings client not configured for credential lookup")
	}
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
	creds := &client.Credentials{
		ClientID: firstNonEmpty(
			credMap[config.HeaderClientID], credMap["client_id"], credMap["FLUTTERWAVE_CLIENT_ID"],
		),
		ClientSecret: firstNonEmpty(
			credMap[config.HeaderClientSecret], credMap["client_secret"], credMap["FLUTTERWAVE_CLIENT_SECRET"],
		),
		PublicKey: firstNonEmpty(
			credMap["public_key"], credMap["FLUTTERWAVE_PUBLIC_KEY"], credMap["FLWPUBK"],
		),
		SecretKey: firstNonEmpty(
			credMap["secret_key"], credMap["FLUTTERWAVE_SECRET_KEY"], credMap["FLWSECK"],
		),
		EncryptionKey: firstNonEmpty(
			credMap["encryption_key"], credMap["FLUTTERWAVE_ENCRYPTION_KEY"],
		),
		WebhookSecret: firstNonEmpty(
			credMap[config.HeaderWebhookSecret], credMap["webhook_secret"], credMap["secret_hash"],
		),
		Environment: firstNonEmpty(
			credMap[config.HeaderEnvironment], credMap["environment"], "sandbox",
		),
		OAuthTokenURL: r.cfg.OAuthTokenURL,
	}
	if strings.HasPrefix(creds.ClientID, "FLWPUBK_") {
		creds.PublicKey = firstNonEmpty(creds.PublicKey, creds.ClientID)
	}
	if strings.HasPrefix(creds.ClientSecret, "FLWSECK_") {
		creds.SecretKey = firstNonEmpty(creds.SecretKey, creds.ClientSecret)
	}
	if strings.EqualFold(creds.Environment, "production") {
		creds.APIBaseURL = r.cfg.ProductionAPIBaseURL
	} else {
		creds.APIBaseURL = r.cfg.SandboxAPIBaseURL
	}
	if client.IsV3Credentials(creds) {
		if creds.SecretKey == "" {
			return nil, errors.New("settings credentials missing FLWSECK secret key")
		}
		return creds, nil
	}
	if creds.ClientID == "" || creds.ClientSecret == "" {
		return nil, errors.New("settings credentials missing client_id/client_secret")
	}
	return creds, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
