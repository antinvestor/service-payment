package queue

import (
	"context"
	"math"
	"net/url"
	"strconv"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/payment/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/config"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/service/client"
	frameEvents "github.com/pitabwire/frame/events"
	"github.com/pitabwire/frame/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
)

type paymentHandler struct {
	credentialResolver
	statusEmitter
	mpesaCli client.MpesaClient
}

// NewPaymentHandler creates a queue worker for handling B2C disbursement payments.
func NewPaymentHandler(
	eventsMan frameEvents.Manager,
	mpesaCli client.MpesaClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.MpesaConfig,
) queue.SubscribeWorker {
	return &paymentHandler{
		credentialResolver: credentialResolver{settingsCli: settingsCli, cfg: cfg},
		statusEmitter:      statusEmitter{eventsMan: eventsMan},
		mpesaCli:           mpesaCli,
	}
}

func (h *paymentHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "mpesa.payment")
	defer logger.Release()
	logger.Debug("queue handler started")

	payment := paymentv1.Payment{}
	if err := proto.Unmarshal(payload, &payment); err != nil {
		logger.WithError(err).Error("failed to unmarshal payment")
		return nil // non-retriable
	}

	paymentID := payment.GetId()
	logger = logger.WithField("payment_id", paymentID)

	creds, err := h.extractCredentials(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "payment",
		})
		return nil
	}

	phoneNumber := payment.GetRecipient().GetContactId()
	amount := formatMoneyAmount(payment.GetAmount())

	commandID := "BusinessPayment"
	if payment.GetExtra() != nil {
		if cmd, ok := payment.GetExtra().GetFields()["command_id"]; ok {
			commandID = cmd.GetStringValue()
		}
	}

	b2cResultURL := appendTenantParams(creds.CallbackURL+"/webhook/mpesa/b2c", headers)
	b2cTimeoutURL := appendTenantParams(creds.CallbackURL+"/webhook/mpesa/b2c/timeout", headers)

	b2cReq := &client.B2CRequest{
		OriginatorConversationID: paymentID,
		InitiatorName:            creds.InitiatorName,
		SecurityCredential:       creds.SecurityCredential,
		CommandID:                commandID,
		Amount:                   amount,
		PartyA:                   creds.Shortcode,
		PartyB:                   phoneNumber,
		Remarks:                  "Payment disbursement",
		QueueTimeOutURL:          b2cTimeoutURL,
		ResultURL:                b2cResultURL,
		Occasion:                 paymentID,
	}

	resp, err := h.mpesaCli.B2CPayment(ctx, creds, b2cReq)
	if err != nil {
		logger.WithError(err).Error("B2C payment failed")
		h.emitStatus(ctx, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "payment",
		})
		return nil
	}

	logger.WithField("conversation_id", resp.ConversationID).Debug("B2C payment initiated")

	h.emitStatus(ctx, paymentID, resp.ConversationID, commonv1.STATUS_IN_PROCESS, map[string]any{
		"conversation_id":            resp.ConversationID,
		"originator_conversation_id": resp.OriginatorConversationID,
		"response_code":              resp.ResponseCode,
		"entity_type":                "payment",
	})

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

// formatMoneyAmount converts a protobuf Money value to a string suitable for M-Pesa (whole KES).
func formatMoneyAmount(amount interface {
	GetUnits() int64
	GetNanos() int32
}) string {
	if amount == nil {
		return "0"
	}
	units := amount.GetUnits()
	nanos := amount.GetNanos()
	total := float64(units) + float64(nanos)/1e9 //nolint:mnd // standard nanos-to-unit conversion
	return strconv.FormatInt(int64(math.Round(total)), 10)
}
