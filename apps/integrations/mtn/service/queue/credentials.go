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

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	settingsv1 "buf.build/gen/go/antinvestor/settingz/protocolbuffers/go/settings/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/mtn/config"
	"github.com/antinvestor/service-payments/apps/integrations/mtn/service/client"
	"github.com/antinvestor/service-payments/pkg/events"
	frameEvents "github.com/pitabwire/frame/v2/events"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/types/known/structpb"
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
		SubscriptionKey: headerOrDefault(headers, config.HeaderSubscriptionKey, cfg.SubscriptionKey),
		APIUser:         headerOrDefault(headers, config.HeaderAPIUser, cfg.APIUser),
		APIKey:          headerOrDefault(headers, config.HeaderAPIKey, cfg.APIKey),
		CallbackURL:     headerOrDefault(headers, config.HeaderCallbackURL, cfg.CallbackURL),
		Currency:        headerOrDefault(headers, config.HeaderCurrency, cfg.Currency),
		Environment:     headerOrDefault(headers, config.HeaderEnvironment, cfg.Environment),
	}

	if creds.SubscriptionKey == "" || creds.APIUser == "" || creds.APIKey == "" {
		return nil, errors.New("missing required MTN MoMo credentials")
	}

	return creds, nil
}

// credentialsFromSettings looks up credentials from the settings service.
func credentialsFromSettings(
	ctx context.Context,
	connection string,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.MtnConfig,
) (*client.MtnCredentials, error) {
	settingReq := &settingsv1.GetRequest{
		Key: &settingsv1.Setting{
			Name:     connection,
			Object:   cfg.SettingsIntegrationName,
			ObjectId: cfg.SettingsIntegrationID,
			Module:   cfg.SettingsIntegrationName,
		},
	}

	settingResp, err := settingsCli.Get(ctx, connect.NewRequest(settingReq))
	if err != nil {
		return nil, fmt.Errorf("settings lookup failed: %w", err)
	}

	var credMap map[string]string
	if err = json.Unmarshal([]byte(settingResp.Msg.GetData().GetValue()), &credMap); err != nil {
		return nil, fmt.Errorf("failed to parse settings credentials: %w", err)
	}

	return &client.MtnCredentials{
		SubscriptionKey: credMap[config.HeaderSubscriptionKey],
		APIUser:         credMap[config.HeaderAPIUser],
		APIKey:          credMap[config.HeaderAPIKey],
		CallbackURL:     credMap[config.HeaderCallbackURL],
		Currency:        credMap[config.HeaderCurrency],
		Environment:     credMap[config.HeaderEnvironment],
	}, nil
}

// emitStatus emits a payment status update event.
func emitStatus(
	ctx context.Context,
	eventsMan frameEvents.Manager,
	id, externalID string,
	status commonv1.STATUS,
	extras map[string]any,
) {
	extra, _ := structpb.NewStruct(extras)
	err := eventsMan.Emit(ctx, events.PaymentStatusUpdateEvent, &commonv1.StatusUpdateRequest{
		Id:         id,
		State:      commonv1.STATE_ACTIVE,
		Status:     status,
		ExternalId: externalID,
		Extras:     extra,
	})
	if err != nil {
		util.Log(ctx).WithError(err).Warn("could not emit status update")
	}
}
