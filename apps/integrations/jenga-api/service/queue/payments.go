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
	"fmt"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/config"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/coreapi"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/events"
	"github.com/pitabwire/frame/queue"
	"github.com/pitabwire/util"
	utilmoney "github.com/pitabwire/util/moneyx"
	"google.golang.org/protobuf/proto"
)

type paymentHandler struct {
	credentialResolver
	statusEmitter
	jengaCli coreapi.JengaApiClient
	metrics  *integrationobs.Metrics
}

// NewPaymentHandler creates a queue worker for handling disbursement payments via Jenga tills-pay.
func NewPaymentHandler(
	eventsMan frameEvents.Manager,
	jengaCli coreapi.JengaApiClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.JengaConfig,
) queue.SubscribeWorker {
	return &paymentHandler{
		credentialResolver: credentialResolver{settingsCli: settingsCli, cfg: cfg},
		statusEmitter:      statusEmitter{eventsMan: eventsMan},
		jengaCli:           jengaCli,
		metrics:            integrationobs.NewMetrics("jenga"),
	}
}

func (h *paymentHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("handler", "jenga.payment")
	defer logger.Release()

	payment := paymentv1.Payment{}
	if err := proto.Unmarshal(payload, &payment); err != nil {
		logger.WithError(err).Error("failed to unmarshal payment protobuf")
		h.metrics.QueueFailed(ctx, "payment", "unmarshal_error")
		return nil // non-retriable
	}

	paymentID := payment.GetId()
	logger = logger.WithFields(map[string]any{
		"payment_id": paymentID,
		"recipient":  payment.GetRecipient().GetContactId(),
	})
	logger.Info("processing payment disbursement")

	jengaCreds, err := h.extractCredentials(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve Jenga credentials")
		h.metrics.QueueFailed(ctx, "payment", "credentials_error")
		h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "payment",
		})
		return nil
	}

	apiCreds := toAPICreds(jengaCreds)

	// Convert protobuf Money to exact decimal string via utilmoney (no float64 intermediate)
	amountDecimal := utilmoney.FromMoney(payment.GetAmount())
	amount := amountDecimal.String()
	currency := payment.GetAmount().GetCurrencyCode()
	if currency == "" {
		currency = "KES"
	}

	tillsReq := models.TillsPayRequest{
		Merchant: models.TillsPayMerchant{
			Till: jengaCreds.MerchantCode,
		},
		Payment: models.TillsPayPayment{
			Ref:      paymentID,
			Amount:   amount,
			Currency: currency,
		},
		Partner: models.TillsPayPartner{
			ID:  payment.GetRecipient().GetContactId(),
			Ref: paymentID,
		},
	}

	token, err := h.jengaCli.GenerateBearerToken(ctx, apiCreds)
	if err != nil {
		logger.WithError(err).Error("failed to generate bearer token for disbursement")
		h.metrics.QueueFailed(ctx, "payment", "credentials_error")
		h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       fmt.Sprintf("generate bearer token: %v", err),
			"entity_type": "payment",
		})
		return nil
	}

	resp, err := h.jengaCli.InitiateTillsPay(ctx, apiCreds, tillsReq, token.AccessToken)
	if err != nil {
		logger.WithError(err).Error("tills pay disbursement failed")
		h.metrics.QueueFailed(ctx, "payment", "provider_error")
		h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       fmt.Sprintf("tills pay: %v", err),
			"entity_type": "payment",
		})
		return nil
	}

	logger.WithFields(map[string]any{
		"transaction_id": resp.TransactionID,
		"merchant_name":  resp.MerchantName,
	}).Info("tills pay disbursement initiated")

	h.emitStatus(ctx, paymentID, resp.TransactionID, commonv1.STATUS_IN_PROCESS, map[string]any{
		"transaction_id": resp.TransactionID,
		"merchant_name":  resp.MerchantName,
		"entity_type":    "payment",
	})

	h.metrics.QueueProcessed(ctx, "payment")
	return nil
}

// toAPICreds converts queue JengaCredentials to the coreapi.Credentials used by the API client.
func toAPICreds(creds *JengaCredentials) *coreapi.Credentials {
	return &coreapi.Credentials{
		MerchantCode:   creds.MerchantCode,
		ConsumerSecret: creds.ConsumerSecret,
		APIKey:         creds.APIKey,
		Environment:    creds.Environment,
		PrivateKeyPath: creds.PrivateKeyPath,
	}
}
