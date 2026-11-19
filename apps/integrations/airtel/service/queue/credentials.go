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
	"strconv"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	settingsv1 "buf.build/gen/go/antinvestor/settingz/protocolbuffers/go/settings/v1"
	"github.com/antinvestor/service-payments/apps/integrations/airtel/config"
	"github.com/antinvestor/service-payments/apps/integrations/airtel/service/client"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/events"
)

const (
	providerName = "airtel"
	kindPayment  = "payment"
	kindPrompt   = "prompt"
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
		req *settingsv1.GetRequest,
	) (*settingsv1.GetResponse, error)
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
		ClientID:        headerOrDefault(headers, config.HeaderClientID, cr.cfg.ClientID),
		ClientSecret:    headerOrDefault(headers, config.HeaderClientSecret, cr.cfg.ClientSecret),
		CallbackURL:     headerOrDefault(headers, config.HeaderCallbackURL, cr.cfg.CallbackURL),
		CountryCode:     headerOrDefault(headers, config.HeaderCountryCode, cr.cfg.CountryCode),
		Currency:        headerOrDefault(headers, config.HeaderCurrency, cr.cfg.Currency),
		Environment:     headerOrDefault(headers, config.HeaderEnvironment, cr.cfg.Environment),
		DisbursementPin: headerOrDefault(headers, config.HeaderDisbursementPin, cr.cfg.DisbursementPin),
		PublicKey:       headerOrDefault(headers, config.HeaderPublicKey, cr.cfg.PublicKey),
	}
	return cr.complete(creds)
}

func (cr *credentialResolver) credentialsFromSettings(
	ctx context.Context,
	connection string,
) (*client.AirtelCredentials, error) {
	if cr.settingsCli == nil {
		return nil, errors.New("settings client not configured; cannot resolve connection credentials")
	}
	settingReq := &settingsv1.GetRequest{
		Key: &settingsv1.Setting{
			Name:     connection,
			Object:   cr.cfg.SettingsIntegrationName,
			ObjectId: cr.cfg.SettingsIntegrationID,
			Module:   cr.cfg.SettingsIntegrationName,
		},
	}

	settingResp, err := cr.settingsCli.Get(ctx, settingReq)
	if err != nil {
		return nil, fmt.Errorf("settings lookup failed: %w", err)
	}

	var credMap map[string]string
	if err = json.Unmarshal([]byte(settingResp.GetData().GetValue()), &credMap); err != nil {
		return nil, fmt.Errorf("failed to parse settings credentials: %w", err)
	}

	return cr.complete(&client.AirtelCredentials{
		ClientID:        credMap[config.HeaderClientID],
		ClientSecret:    credMap[config.HeaderClientSecret],
		CallbackURL:     credMap[config.HeaderCallbackURL],
		CountryCode:     credMap[config.HeaderCountryCode],
		Currency:        credMap[config.HeaderCurrency],
		Environment:     credMap[config.HeaderEnvironment],
		DisbursementPin: credMap[config.HeaderDisbursementPin],
		PublicKey:       credMap[config.HeaderPublicKey],
	})
}

// complete fills configuration defaults and checks the credentials are usable.
func (cr *credentialResolver) complete(creds *client.AirtelCredentials) (*client.AirtelCredentials, error) {
	if creds.Environment == "" {
		creds.Environment = cr.cfg.Environment
	}
	if creds.CountryCode == "" {
		creds.CountryCode = cr.cfg.CountryCode
	}
	if creds.Currency == "" {
		creds.Currency = cr.cfg.Currency
	}
	if creds.ClientID == "" || creds.ClientSecret == "" {
		return nil, errors.New("missing required Airtel Money credentials")
	}
	return creds, nil
}

// withRequestBinding adds what the callback handler needs to re-query Airtel:
// the requested amount/currency and the credential connection key.
func withRequestBinding(extras map[string]any, amount, currency string, headers map[string]string) map[string]any {
	extras[client.ExtraRequestedAmount] = amount
	extras[client.ExtraRequestedCurrency] = currency
	if connection := headers[config.HeaderConnectionCredentials]; connection != "" {
		extras[client.ExtraCredentialsConnection] = connection
	}
	return extras
}

// airtelAmount renders a Money for Airtel, which moves whole currency units
// only; a fractional amount is refused rather than rounded.
func airtelAmount(m *commonv1.Money) (string, error) {
	units, err := collection.WholeUnits(m)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(units, 10), nil
}

// requestCurrency is the prompt/payment currency, else the credentials'.
func requestCurrency(m *commonv1.Money, creds *client.AirtelCredentials) string {
	if c := strings.ToUpper(strings.TrimSpace(m.GetCurrencyCode())); c != "" {
		return c
	}
	return strings.ToUpper(strings.TrimSpace(creds.Currency))
}

// mapAirtelStatus maps Airtel transaction status codes to the platform status.
func mapAirtelStatus(statusCode string) commonv1.STATUS {
	switch strings.ToUpper(strings.TrimSpace(statusCode)) {
	case "TS":
		return commonv1.STATUS_SUCCESSFUL
	case "TF":
		return commonv1.STATUS_FAILED
	case "TIP", "TA":
		return commonv1.STATUS_IN_PROCESS
	default:
		return commonv1.STATUS_UNKNOWN
	}
}

// statusExtras describes the state Airtel holds for a transaction.
func statusExtras(extras map[string]any, t *client.StatusResponse) map[string]any {
	txn := t.Data.Transaction
	extras[client.ExtraTransactionID] = txn.ID
	extras["status_code"] = txn.Status
	extras["message"] = txn.Message
	if txn.AirtelMoneyID != "" {
		extras["airtel_money_id"] = string(txn.AirtelMoneyID)
	}
	return extras
}

// emitKnown reports the state Airtel already holds for an entity.
func emitKnown(
	ctx context.Context,
	emitter *events.StatusEmitter,
	entityType, id string,
	st *client.StatusResponse,
	extras map[string]any,
) error {
	status := mapAirtelStatus(st.Data.Transaction.Status)
	extras = statusExtras(extras, st)
	if status == commonv1.STATUS_FAILED {
		reason := st.Data.Transaction.Message
		if reason == "" {
			reason = "airtel reported the transaction failed"
		}
		return emitter.Fail(ctx, entityType, id, reason, extras)
	}
	return emitter.Emit(ctx, entityType, id, string(st.Data.Transaction.AirtelMoneyID), status, extras)
}

// NewCredentialResolver returns a resolver for credentials outside the queue
// workers (e.g. to re-query Airtel from a callback): the settings connection
// when known, otherwise the service configuration.
func NewCredentialResolver(
	settingsCli settingsClient,
	cfg *config.AirtelConfig,
) func(ctx context.Context, connection string) (*client.AirtelCredentials, error) {
	cr := &credentialResolver{settingsCli: settingsCli, cfg: cfg}
	return func(ctx context.Context, connection string) (*client.AirtelCredentials, error) {
		headers := map[string]string{}
		if connection != "" {
			headers[config.HeaderConnectionCredentials] = connection
		}
		return cr.extractCredentials(ctx, headers)
	}
}

func headerOrDefault(headers map[string]string, key, fallback string) string {
	if v, ok := headers[key]; ok && v != "" {
		return v
	}
	return fallback
}
