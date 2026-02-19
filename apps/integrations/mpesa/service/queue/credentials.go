package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	settingsv1 "buf.build/gen/go/antinvestor/settingz/protocolbuffers/go/settings/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/config"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/service/client"
)

// credentialResolver provides shared credential extraction logic for queue handlers.
type credentialResolver struct {
	settingsCli settingsv1connect.SettingsServiceClient
	cfg         *config.MpesaConfig
}

func (r *credentialResolver) extractCredentials(
	ctx context.Context,
	headers map[string]string,
) (*client.MpesaCredentials, error) {
	// Try settings service lookup first
	connection, ok := headers[config.HeaderConnectionCredentials]
	if ok && connection != "" {
		return r.credentialsFromSettings(ctx, connection)
	}

	// Try direct headers, fall back to config
	creds := &client.MpesaCredentials{
		ConsumerKey:        headerOrDefault(headers, config.HeaderConsumerKey, r.cfg.ConsumerKey),
		ConsumerSecret:     headerOrDefault(headers, config.HeaderConsumerSecret, r.cfg.ConsumerSecret),
		Shortcode:          headerOrDefault(headers, config.HeaderShortcode, r.cfg.Shortcode),
		Passkey:            headerOrDefault(headers, config.HeaderPasskey, r.cfg.Passkey),
		CallbackURL:        headerOrDefault(headers, config.HeaderCallbackURL, r.cfg.CallbackURL),
		InitiatorName:      headerOrDefault(headers, config.HeaderInitiatorName, r.cfg.InitiatorName),
		SecurityCredential: headerOrDefault(headers, config.HeaderSecurityCredential, r.cfg.SecurityCredential),
		Environment:        headerOrDefault(headers, config.HeaderEnvironment, r.cfg.Environment),
	}

	if creds.ConsumerKey == "" || creds.ConsumerSecret == "" {
		return nil, errors.New("missing required M-Pesa credentials (consumer_key, consumer_secret)")
	}

	return creds, nil
}

func (r *credentialResolver) credentialsFromSettings(
	ctx context.Context,
	connection string,
) (*client.MpesaCredentials, error) {
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
		return nil, fmt.Errorf("failed to parse settings credentials: %w", err)
	}

	creds := &client.MpesaCredentials{
		ConsumerKey:        credMap[config.HeaderConsumerKey],
		ConsumerSecret:     credMap[config.HeaderConsumerSecret],
		Shortcode:          credMap[config.HeaderShortcode],
		Passkey:            credMap[config.HeaderPasskey],
		CallbackURL:        credMap[config.HeaderCallbackURL],
		InitiatorName:      credMap[config.HeaderInitiatorName],
		SecurityCredential: credMap[config.HeaderSecurityCredential],
		Environment:        credMap[config.HeaderEnvironment],
	}

	return creds, nil
}

func headerOrDefault(headers map[string]string, key, fallback string) string {
	if v, ok := headers[key]; ok && v != "" {
		return v
	}
	return fallback
}
