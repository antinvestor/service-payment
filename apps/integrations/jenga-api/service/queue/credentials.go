package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	settingsv1 "buf.build/gen/go/antinvestor/settingz/protocolbuffers/go/settings/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/config"
)

// JengaCredentials holds per-request credentials for Jenga API calls.
type JengaCredentials struct {
	MerchantCode   string
	ConsumerSecret string
	APIKey         string
	CallbackURL    string
	Environment    string
	PrivateKeyPath string
}

// credentialResolver provides shared credential extraction logic for queue handlers.
type credentialResolver struct {
	settingsCli settingsv1connect.SettingsServiceClient
	cfg         *config.JengaConfig
}

func (r *credentialResolver) extractCredentials(
	ctx context.Context,
	headers map[string]string,
) (*JengaCredentials, error) {
	// Try settings service lookup first (per-tenant)
	connection, ok := headers[config.HeaderConnectionCredentials]
	if ok && connection != "" && r.settingsCli != nil {
		return r.credentialsFromSettings(ctx, connection)
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
