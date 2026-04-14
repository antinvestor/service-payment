package queue

import (
	"context"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/airtel/config"
	"github.com/antinvestor/service-payments/apps/integrations/airtel/service/client"
	"github.com/antinvestor/service-payments/pkg/events"
	frameEvents "github.com/pitabwire/frame/events"
	"github.com/pitabwire/frame/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type promptHandler struct {
	credentialResolver
	eventsMan frameEvents.Manager
	airtelCli client.AirtelClient
}

// NewPromptHandler creates a queue worker for handling USSD push collection prompts.
func NewPromptHandler(
	eventsMan frameEvents.Manager,
	airtelCli client.AirtelClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.AirtelConfig,
) queue.SubscribeWorker {
	return &promptHandler{
		credentialResolver: credentialResolver{
			settingsCli: settingsCli,
			cfg:         cfg,
		},
		eventsMan: eventsMan,
		airtelCli: airtelCli,
	}
}

func (h *promptHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "airtel.prompt")
	defer logger.Release()
	logger.Debug("queue handler started")

	prompt := paymentv1.InitiatePromptRequest{}
	if err := proto.Unmarshal(payload, &prompt); err != nil {
		logger.WithError(err).Error("failed to unmarshal prompt")
		return nil
	}

	promptID := prompt.GetId()
	logger = logger.WithField("prompt_id", promptID)

	creds, err := h.extractCredentials(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "prompt",
		})
		return nil
	}

	phoneNumber := prompt.GetRecipient().GetContactId()
	if phoneNumber == "" {
		phoneNumber = prompt.GetSource().GetContactId()
	}

	amount := formatMoneyAmount(prompt.GetAmount())

	currency := creds.Currency
	if prompt.GetAmount() != nil && prompt.GetAmount().GetCurrencyCode() != "" {
		currency = prompt.GetAmount().GetCurrencyCode()
	}

	req := &client.CollectionRequest{
		Reference:   promptID,
		PhoneNumber: phoneNumber,
		Amount:      amount,
		Currency:    currency,
		CountryCode: creds.CountryCode,
		CallbackURL: appendTenantParams(creds.CallbackURL+"/webhook/airtel/collection", headers),
	}

	resp, err := h.airtelCli.CollectionPush(ctx, creds, req)
	if err != nil {
		logger.WithError(err).Error("collection push failed")
		h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "prompt",
		})
		return nil
	}

	externalID := resp.Data.Transaction.ID
	logger.WithField("transaction_id", externalID).Debug("collection push initiated")

	h.emitStatus(ctx, promptID, externalID, commonv1.STATUS_IN_PROCESS, map[string]any{
		"transaction_id": externalID,
		"status_code":    resp.Status.Code,
		"message":        resp.Status.Message,
		"entity_type":    "prompt",
	})

	return nil
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
