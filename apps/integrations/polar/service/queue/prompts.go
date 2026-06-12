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
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	settingsv1 "buf.build/gen/go/antinvestor/settingz/protocolbuffers/go/settings/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/polar/config"
	"github.com/antinvestor/service-payments/apps/integrations/polar/service/client"
	"github.com/antinvestor/service-payments/pkg/events"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/events"
	"github.com/pitabwire/frame/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type promptHandler struct {
	eventsMan   frameEvents.Manager
	polarCli    client.PolarClient
	settingsCli settingsv1connect.SettingsServiceClient
	cfg         *config.PolarConfig
	metrics     *integrationobs.Metrics
}

// NewPromptHandler creates a queue worker for handling Polar checkout session prompts.
func NewPromptHandler(
	eventsMan frameEvents.Manager,
	polarCli client.PolarClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.PolarConfig,
) queue.SubscribeWorker {
	return &promptHandler{
		eventsMan:   eventsMan,
		polarCli:    polarCli,
		settingsCli: settingsCli,
		cfg:         cfg,
		metrics:     integrationobs.NewMetrics("polar"),
	}
}

func (h *promptHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "polar.prompt")
	defer logger.Release()
	logger.Debug("queue handler started")

	prompt := paymentv1.InitiatePromptRequest{}
	if err := proto.Unmarshal(payload, &prompt); err != nil {
		logger.WithError(err).Error("failed to unmarshal prompt")
		h.metrics.QueueFailed(ctx, "prompt", "unmarshal_error")
		return nil
	}

	promptID := prompt.GetId()
	logger = logger.WithField("prompt_id", promptID)

	creds, err := h.extractCredentials(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		h.metrics.QueueFailed(ctx, "prompt", "credentials_error")
		h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "prompt",
		})
		return nil
	}

	var amount int64
	if prompt.GetAmount() != nil {
		amount = moneyToCents(prompt.GetAmount().GetUnits(), prompt.GetAmount().GetNanos())
	}

	currency := "usd"
	if prompt.GetAmount() != nil && prompt.GetAmount().GetCurrencyCode() != "" {
		currency = prompt.GetAmount().GetCurrencyCode()
	}

	productID := ""
	customerEmail := ""
	successURL := ""
	if prompt.GetExtra() != nil {
		fields := prompt.GetExtra().GetFields()
		if v, ok := fields["product_id"]; ok {
			productID = v.GetStringValue()
		}
		if v, ok := fields["customer_email"]; ok {
			customerEmail = v.GetStringValue()
		}
		if v, ok := fields["success_url"]; ok {
			successURL = v.GetStringValue()
		}
	}

	checkoutMetadata := map[string]string{
		"prompt_id": promptID,
	}
	// Include tenant/partition in Polar metadata so they round-trip via webhooks
	if v := headers["tenant_id"]; v != "" {
		checkoutMetadata["tenant_id"] = v
	}
	if v := headers["partition_id"]; v != "" {
		checkoutMetadata["partition_id"] = v
	}

	req := &client.CheckoutRequest{
		ProductID:     productID,
		CustomerEmail: customerEmail,
		Amount:        amount,
		Currency:      currency,
		SuccessURL:    successURL,
		Metadata:      checkoutMetadata,
	}

	resp, err := h.polarCli.CreateCheckout(ctx, creds, req)
	if err != nil {
		logger.WithError(err).Error("create checkout failed")
		h.metrics.QueueFailed(ctx, "prompt", "provider_error")
		h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "prompt",
		})
		return nil
	}

	logger.WithField("checkout_id", resp.ID).Debug("checkout session created")

	h.emitStatus(ctx, promptID, resp.ID, commonv1.STATUS_IN_PROCESS, map[string]any{
		"checkout_id":  resp.ID,
		"checkout_url": resp.URL,
		"polar_status": resp.Status,
		"entity_type":  "prompt",
	})

	h.metrics.QueueProcessed(ctx, "prompt")
	return nil
}

func (h *promptHandler) extractCredentials(
	ctx context.Context,
	headers map[string]string,
) (*client.PolarCredentials, error) {
	connection, ok := headers[config.HeaderConnectionCredentials]
	if ok && connection != "" {
		return h.credentialsFromSettings(ctx, connection)
	}

	creds := &client.PolarCredentials{
		APIKey:         headerOrDefault(headers, config.HeaderAPIKey, h.cfg.APIKey),
		WebhookSecret:  headerOrDefault(headers, config.HeaderWebhookSecret, h.cfg.WebhookSecret),
		OrganizationID: headerOrDefault(headers, config.HeaderOrganizationID, h.cfg.OrganizationID),
		Environment:    headerOrDefault(headers, config.HeaderEnvironment, h.cfg.Environment),
	}

	if creds.APIKey == "" {
		return nil, errors.New("missing required Polar API key")
	}

	return creds, nil
}

func (h *promptHandler) credentialsFromSettings(
	ctx context.Context,
	connection string,
) (*client.PolarCredentials, error) {
	settingReq := &settingsv1.GetRequest{
		Key: &settingsv1.Setting{
			Name:     connection,
			Object:   h.cfg.SettingsIntegrationName,
			ObjectId: h.cfg.SettingsIntegrationID,
			Module:   h.cfg.SettingsIntegrationName,
		},
	}

	settingResp, err := h.settingsCli.Get(ctx, connect.NewRequest(settingReq))
	if err != nil {
		return nil, fmt.Errorf("settings lookup failed: %w", err)
	}

	var credMap map[string]string
	if err = json.Unmarshal([]byte(settingResp.Msg.GetData().GetValue()), &credMap); err != nil {
		return nil, fmt.Errorf("failed to parse settings credentials: %w", err)
	}

	return &client.PolarCredentials{
		APIKey:         credMap[config.HeaderAPIKey],
		WebhookSecret:  credMap[config.HeaderWebhookSecret],
		OrganizationID: credMap[config.HeaderOrganizationID],
		Environment:    credMap[config.HeaderEnvironment],
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

func (h *promptHandler) emitStatus(
	ctx context.Context,
	id, externalID string,
	status commonv1.STATUS,
	extras map[string]any,
) {
	extra, _ := structpb.NewStruct(extras)
	err := h.eventsMan.Emit(ctx, events.PaymentStatusUpdateEvent, &commonv1.StatusUpdateRequest{
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
