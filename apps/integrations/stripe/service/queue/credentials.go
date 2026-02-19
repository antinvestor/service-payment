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
	"github.com/antinvestor/service-payments/apps/integrations/stripe/config"
	"github.com/antinvestor/service-payments/apps/integrations/stripe/service/client"
	"github.com/antinvestor/service-payments/internal/events"
	frameEvents "github.com/pitabwire/frame/events"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/types/known/structpb"
)

func extractCredentials(
	ctx context.Context,
	headers map[string]string,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.StripeConfig,
) (*client.StripeCredentials, error) {
	connection, ok := headers[config.HeaderConnectionCredentials]
	if ok && connection != "" {
		return credentialsFromSettings(ctx, connection, settingsCli, cfg)
	}

	creds := &client.StripeCredentials{
		APIKey:        headerOrDefault(headers, config.HeaderAPIKey, cfg.APIKey),
		WebhookSecret: headerOrDefault(headers, config.HeaderWebhookSecret, cfg.WebhookSecret),
	}

	if creds.APIKey == "" {
		return nil, errors.New("missing required Stripe API key")
	}

	return creds, nil
}

func credentialsFromSettings(
	ctx context.Context,
	connection string,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.StripeConfig,
) (*client.StripeCredentials, error) {
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

	return &client.StripeCredentials{
		APIKey:        credMap[config.HeaderAPIKey],
		WebhookSecret: credMap[config.HeaderWebhookSecret],
	}, nil
}

const (
	centsPerUnit = 100
	nanosPerCent = 10000000
)

// moneyToCents converts google.type.Money units+nanos to smallest currency unit (cents).
func moneyToCents(units int64, nanos int32) int64 {
	return units*centsPerUnit + int64(nanos/nanosPerCent)
}

func headerOrDefault(headers map[string]string, key, fallback string) string {
	if v, ok := headers[key]; ok && v != "" {
		return v
	}
	return fallback
}

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
