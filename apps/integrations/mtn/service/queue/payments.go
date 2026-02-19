package queue

import (
	"context"
	"math"
	"net/url"
	"strconv"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/payment/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/mtn/config"
	"github.com/antinvestor/service-payments/apps/integrations/mtn/service/client"
	"github.com/google/uuid"
	frameEvents "github.com/pitabwire/frame/events"
	"github.com/pitabwire/frame/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
)

type paymentHandler struct {
	eventsMan   frameEvents.Manager
	mtnCli      client.MtnClient
	settingsCli settingsv1connect.SettingsServiceClient
	cfg         *config.MtnConfig
}

// NewPaymentHandler creates a queue worker for handling disbursement transfer payments.
func NewPaymentHandler(
	eventsMan frameEvents.Manager,
	mtnCli client.MtnClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.MtnConfig,
) queue.SubscribeWorker {
	return &paymentHandler{
		eventsMan:   eventsMan,
		mtnCli:      mtnCli,
		settingsCli: settingsCli,
		cfg:         cfg,
	}
}

func (h *paymentHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "mtn.payment")
	defer logger.Release()
	logger.Debug("queue handler started")

	payment := paymentv1.Payment{}
	if err := proto.Unmarshal(payload, &payment); err != nil {
		logger.WithError(err).Error("failed to unmarshal payment")
		return nil
	}

	paymentID := payment.GetId()
	logger = logger.WithField("payment_id", paymentID)

	creds, err := extractCredentials(ctx, headers, h.settingsCli, h.cfg)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		emitStatus(ctx, h.eventsMan, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
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

	referenceID := uuid.NewString()

	req := &client.TransferRequest{
		ReferenceID: referenceID,
		Amount:      amount,
		Currency:    currency,
		ExternalID:  paymentID,
		Payee: client.Party{
			PartyIDType: "MSISDN",
			PartyID:     phoneNumber,
		},
		PayerMessage: "Payment disbursement",
		PayeeNote:    paymentID,
		CallbackURL:  appendTenantParams(creds.CallbackURL+"/webhook/mtn/disbursement", headers),
	}

	err = h.mtnCli.Transfer(ctx, creds, req)
	if err != nil {
		logger.WithError(err).Error("transfer failed")
		emitStatus(ctx, h.eventsMan, paymentID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "payment",
		})
		return nil
	}

	logger.WithField("reference_id", referenceID).Debug("transfer initiated")

	emitStatus(ctx, h.eventsMan, paymentID, referenceID, commonv1.STATUS_IN_PROCESS, map[string]any{
		"reference_id": referenceID,
		"entity_type":  "payment",
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

func headerOrDefault(headers map[string]string, key, fallback string) string {
	if v, ok := headers[key]; ok && v != "" {
		return v
	}
	return fallback
}
