package queue

import (
	"context"
	"fmt"
	"strconv"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/config"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/coreapi"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	frameEvents "github.com/pitabwire/frame/events"
	"github.com/pitabwire/frame/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
)

type paymentHandler struct {
	credentialResolver
	statusEmitter
	jengaCli coreapi.JengaApiClient
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
	}
}

func (h *paymentHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("handler", "jenga.payment")
	defer logger.Release()

	payment := paymentv1.Payment{}
	if err := proto.Unmarshal(payload, &payment); err != nil {
		logger.WithError(err).Error("failed to unmarshal payment protobuf")
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
		h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "payment",
		})
		return nil
	}

	apiCreds := toAPICreds(jengaCreds)

	// Format amount preserving sub-unit precision
	amount := formatMoneyAmount(payment.GetAmount())
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
		h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       fmt.Sprintf("generate bearer token: %v", err),
			"entity_type": "payment",
		})
		return nil
	}

	resp, err := h.jengaCli.InitiateTillsPay(ctx, apiCreds, tillsReq, token.AccessToken)
	if err != nil {
		logger.WithError(err).Error("tills pay disbursement failed")
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

	return nil
}

// formatMoneyAmount converts a protobuf Money value to a decimal string preserving sub-unit precision.
func formatMoneyAmount(amount interface {
	GetUnits() int64
	GetNanos() int32
},
) string {
	if amount == nil {
		return "0"
	}
	units := amount.GetUnits()
	nanos := amount.GetNanos()
	total := float64(units) + float64(nanos)/1e9 //nolint:mnd // nano conversion factor
	return strconv.FormatFloat(total, 'f', 2, 64)
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
