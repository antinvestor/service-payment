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

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/v2/settings/v1/settingsv1connect"
	settingsv1 "buf.build/gen/go/antinvestor/settingz/protocolbuffers/go/settings/v1"
	"github.com/antinvestor/service-payments/apps/integrations/mtn/config"
	"github.com/antinvestor/service-payments/apps/integrations/mtn/service/client"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/events"
)

// extractCredentials resolves MTN MoMo credentials from headers or settings.
func extractCredentials(
	ctx context.Context,
	headers map[string]string,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.MtnConfig,
) (*client.MtnCredentials, error) {
	connection, ok := headers[config.HeaderConnectionCredentials]
	if ok && connection != "" {
		return credentialsFromSettings(ctx, connection, settingsCli, cfg)
	}

	creds := &client.MtnCredentials{
		SubscriptionKey:   headerOrDefault(headers, config.HeaderSubscriptionKey, cfg.SubscriptionKey),
		APIUser:           headerOrDefault(headers, config.HeaderAPIUser, cfg.APIUser),
		APIKey:            headerOrDefault(headers, config.HeaderAPIKey, cfg.APIKey),
		CallbackURL:       headerOrDefault(headers, config.HeaderCallbackURL, cfg.CallbackURL),
		Currency:          headerOrDefault(headers, config.HeaderCurrency, cfg.Currency),
		Environment:       headerOrDefault(headers, config.HeaderEnvironment, cfg.Environment),
		TargetEnvironment: headerOrDefault(headers, config.HeaderTargetEnvironment, cfg.TargetEnvironment),
	}
	return completeCredentials(creds, cfg)
}

// completeCredentials fills environment-derived defaults and checks the
// credentials are usable.
func completeCredentials(creds *client.MtnCredentials, cfg *config.MtnConfig) (*client.MtnCredentials, error) {
	if creds.Environment == "" {
		creds.Environment = cfg.Environment
	}
	if creds.Currency == "" {
		creds.Currency = cfg.Currency
	}
	if creds.SubscriptionKey == "" || creds.APIUser == "" || creds.APIKey == "" {
		return nil, errors.New("missing required MTN MoMo credentials")
	}
	target, err := config.ResolveTargetEnvironment(creds.Environment, creds.TargetEnvironment)
	if err != nil {
		return nil, err
	}
	creds.TargetEnvironment = target
	return creds, nil
}

// credentialsFromSettings looks up credentials from the settings service.
func credentialsFromSettings(
	ctx context.Context,
	connection string,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.MtnConfig,
) (*client.MtnCredentials, error) {
	if settingsCli == nil {
		return nil, errors.New("settings client not configured; cannot resolve connection credentials")
	}
	settingReq := &settingsv1.GetRequest{
		Key: &settingsv1.Setting{
			Name:     connection,
			Object:   cfg.SettingsIntegrationName,
			ObjectId: cfg.SettingsIntegrationID,
			Module:   cfg.SettingsIntegrationName,
		},
	}

	settingResp, err := settingsCli.Get(ctx, settingReq)
	if err != nil {
		return nil, fmt.Errorf("settings lookup failed: %w", err)
	}

	var credMap map[string]string
	if err = json.Unmarshal([]byte(settingResp.GetData().GetValue()), &credMap); err != nil {
		return nil, fmt.Errorf("failed to parse settings credentials: %w", err)
	}

	return completeCredentials(&client.MtnCredentials{
		SubscriptionKey:   credMap[config.HeaderSubscriptionKey],
		APIUser:           credMap[config.HeaderAPIUser],
		APIKey:            credMap[config.HeaderAPIKey],
		CallbackURL:       credMap[config.HeaderCallbackURL],
		Currency:          credMap[config.HeaderCurrency],
		Environment:       credMap[config.HeaderEnvironment],
		TargetEnvironment: credMap[config.HeaderTargetEnvironment],
	}, cfg)
}

// requestExtras is what the callback handler needs to re-query MTN and check
// the amount: the reference id, the amount/currency sent and the credential
// connection key.
func requestExtras(entityType, referenceID, amount, currency string, headers map[string]string) map[string]any {
	extras := map[string]any{
		client.ExtraReferenceID:       referenceID,
		client.ExtraRequestedAmount:   amount,
		client.ExtraRequestedCurrency: currency,
		events.ExtraEntityType:        entityType,
	}
	if connection := headers[config.HeaderConnectionCredentials]; connection != "" {
		extras[client.ExtraCredentialsConnection] = connection
	}
	return extras
}

// mtnAmount renders a Money for MTN: whole units when there is no fraction
// ("5000"), else the currency's decimals ("10.50"). An amount finer than the
// currency allows is refused rather than rounded into a different charge.
func mtnAmount(m *commonv1.Money) (string, error) {
	if _, err := collection.MinorUnits(m); err != nil {
		return "", err
	}
	if m.GetNanos() == 0 {
		return fmt.Sprintf("%d", m.GetUnits()), nil
	}
	return collection.DecimalString(m), nil
}

// requestCurrency is the prompt/payment currency, else the credentials'.
func requestCurrency(m *commonv1.Money, creds *client.MtnCredentials) string {
	if c := strings.ToUpper(strings.TrimSpace(m.GetCurrencyCode())); c != "" {
		return c
	}
	return strings.ToUpper(strings.TrimSpace(creds.Currency))
}

// mapMtnStatus maps MTN's transaction status to the platform status.
func mapMtnStatus(status string) commonv1.STATUS {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "SUCCESSFUL":
		return commonv1.STATUS_SUCCESSFUL
	case "FAILED":
		return commonv1.STATUS_FAILED
	case "PENDING":
		return commonv1.STATUS_IN_PROCESS
	default:
		return commonv1.STATUS_UNKNOWN
	}
}

// NewCredentialResolver returns a resolver for credentials outside the queue
// workers (e.g. to re-query MTN from a callback): the settings connection
// when known, otherwise the service configuration.
func NewCredentialResolver(
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.MtnConfig,
) func(ctx context.Context, connection string) (*client.MtnCredentials, error) {
	return func(ctx context.Context, connection string) (*client.MtnCredentials, error) {
		headers := map[string]string{}
		if connection != "" {
			if settingsCli == nil {
				return nil, errors.New("settings client unavailable for credential connection lookup")
			}
			headers[config.HeaderConnectionCredentials] = connection
		}
		return extractCredentials(ctx, headers, settingsCli, cfg)
	}
}
