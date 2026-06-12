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
	"math"
	"net/url"
	"strconv"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/airtel/config"
	"github.com/antinvestor/service-payments/apps/integrations/airtel/service/client"
	"github.com/antinvestor/service-payments/pkg/events"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/events"
	"github.com/pitabwire/frame/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type paymentHandler struct {
	credentialResolver
	eventsMan frameEvents.Manager
	airtelCli client.AirtelClient
	metrics   *integrationobs.Metrics
}

// NewPaymentHandler creates a queue worker for handling disbursement payments.
func NewPaymentHandler(
	eventsMan frameEvents.Manager,
	airtelCli client.AirtelClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.AirtelConfig,
) queue.SubscribeWorker {
	return &paymentHandler{
		credentialResolver: credentialResolver{
			settingsCli: settingsCli,
			cfg:         cfg,
		},
		eventsMan: eventsMan,
		airtelCli: airtelCli,
		metrics:   integrationobs.NewMetrics("airtel"),
	}
}

func (h *paymentHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "airtel.payment")
	defer logger.Release()
	logger.Debug("queue handler started")

	payment := paymentv1.Payment{}
	if err := proto.Unmarshal(payload, &payment); err != nil {
		logger.WithError(err).Error("failed to unmarshal payment")
		h.metrics.QueueFailed(ctx, "payment", "unmarshal_error")
		return nil
	}

	paymentID := payment.GetId()
	logger = logger.WithField("payment_id", paymentID)

	creds, err := h.extractCredentials(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		h.metrics.QueueFailed(ctx, "payment", "credentials_error")
		h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "payment",
		})
		return nil
	}

	phoneNumber := payment.GetRecipient().GetContactId()
	amount := formatMoneyAmount(payment.GetAmount())

	currency := creds.Currency
	if payment.GetAmount() != nil && payment.GetAmount().GetCurrencyCode() != "" {
		currency = payment.GetAmount().GetCurrencyCode()
	}

	req := &client.DisbursementRequest{
		Reference:   paymentID,
		PhoneNumber: phoneNumber,
		Amount:      amount,
		Currency:    currency,
		CountryCode: creds.CountryCode,
	}

	resp, err := h.airtelCli.Disburse(ctx, creds, req)
	if err != nil {
		logger.WithError(err).Error("disbursement failed")
		h.metrics.QueueFailed(ctx, "payment", "provider_error")
		h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "payment",
		})
		return nil
	}

	externalID := resp.Data.Transaction.ID
	logger.WithField("transaction_id", externalID).Debug("disbursement initiated")

	h.emitStatus(ctx, paymentID, externalID, commonv1.STATUS_IN_PROCESS, map[string]any{
		"transaction_id": externalID,
		"reference_id":   resp.Data.Transaction.ReferenceID,
		"status_code":    resp.Status.Code,
		"message":        resp.Status.Message,
		"entity_type":    "payment",
	})

	h.metrics.QueueProcessed(ctx, "payment")
	return nil
}

// appendTenantParams appends tenant_id and partition_id query params to a callback URL.
func appendTenantParams(callbackURL string, headers map[string]string) string {
	tenantID := headers["tenant_id"]
	partitionID := headers["partition_id"]
	if tenantID == "" && partitionID == "" {
		return callbackURL
	}
	u, err := url.Parse(callbackURL)
	if err != nil {
		return callbackURL
	}
	q := u.Query()
	if tenantID != "" {
		q.Set("tenant_id", tenantID)
	}
	if partitionID != "" {
		q.Set("partition_id", partitionID)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func formatMoneyAmount(amount interface {
	GetUnits() int64
	GetNanos() int32
}) string {
	if amount == nil {
		return "0"
	}
	units := amount.GetUnits()
	nanos := amount.GetNanos()
	total := float64(units) + float64(nanos)/1e9 //nolint:mnd // nanos to units conversion
	return strconv.FormatInt(int64(math.Round(total)), 10)
}

func headerOrDefault(headers map[string]string, key, fallback string) string {
	if v, ok := headers[key]; ok && v != "" {
		return v
	}
	return fallback
}

func (h *paymentHandler) emitStatus(
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
