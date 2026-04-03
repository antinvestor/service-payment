package handlers

import (
	"context"
	"io"
	"net/http"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/polar/config"
	"github.com/antinvestor/service-payments/apps/integrations/polar/service/client"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/security"
	"github.com/pitabwire/util"
)

// PolarWebhookServer handles Polar.sh webhook events.
type PolarWebhookServer struct {
	paymentCli    paymentv1connect.PaymentServiceClient
	polarCli      client.PolarClient
	webhookSecret string
}

// NewPolarWebhookServer creates a new webhook server.
func NewPolarWebhookServer(
	paymentCli paymentv1connect.PaymentServiceClient,
	polarCli client.PolarClient,
	cfg *config.PolarConfig,
) *PolarWebhookServer {
	return &PolarWebhookServer{
		paymentCli:    paymentCli,
		polarCli:      polarCli,
		webhookSecret: cfg.WebhookSecret,
	}
}

// NewRouterV1 creates the HTTP routes for Polar webhooks.
func (s *PolarWebhookServer) NewRouterV1() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/polar", s.HandleWebhook)
	return mux
}

// HandleWebhook processes Polar webhook events.
func (s *PolarWebhookServer) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := util.Log(ctx).WithField("type", "polar.webhook")
	defer logger.Release()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.WithError(err).Error("failed to read webhook body")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	headers := map[string]string{
		"webhook-signature": r.Header.Get("Webhook-Signature"),
		"webhook-id":        r.Header.Get("Webhook-Id"),
		"webhook-timestamp": r.Header.Get("Webhook-Timestamp"),
	}

	event, err := s.polarCli.VerifyWebhookSignature(body, headers, s.webhookSecret)
	if err != nil {
		logger.WithError(err).Error("webhook signature verification failed")
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	logger = logger.WithField("event_type", event.Type)
	logger.Debug("processing Polar webhook event")

	var statusErr error
	switch event.Type {
	case "checkout.created":
		statusErr = s.handleCheckoutCreated(ctx, logger, event)
	case "checkout.updated":
		statusErr = s.handleCheckoutUpdated(ctx, logger, event)
	case "subscription.created":
		statusErr = s.handleSubscriptionCreated(ctx, logger, event)
	case "subscription.updated":
		statusErr = s.handleSubscriptionUpdated(ctx, logger, event)
	default:
		logger.Debug("unhandled event type, skipping")
	}

	if statusErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *PolarWebhookServer) handleCheckoutCreated(
	ctx context.Context,
	logger *util.LogEntry,
	event *client.WebhookEvent,
) error {
	checkoutID := getStringField(event.Data, "id")
	metadata := getMetadata(event.Data)
	ctx = injectTenantContext(ctx, metadata)

	promptID := metadata["prompt_id"]
	if promptID == "" {
		promptID = checkoutID
	}

	extras := data.JSONMap{
		"polar_event_type": event.Type,
		"checkout_id":      checkoutID,
		"entity_type":      "prompt",
	}

	statusReq := &commonv1.StatusUpdateRequest{
		Id:         promptID,
		State:      commonv1.STATE_ACTIVE,
		Status:     commonv1.STATUS_IN_PROCESS,
		ExternalId: checkoutID,
		Extras:     extras.ToProtoStruct(),
	}

	if _, err := s.paymentCli.StatusUpdate(ctx, connect.NewRequest(statusReq)); err != nil {
		logger.WithError(err).Error("could not update payment status")
		return err
	}
	return nil
}

func (s *PolarWebhookServer) handleCheckoutUpdated(
	ctx context.Context,
	logger *util.LogEntry,
	event *client.WebhookEvent,
) error {
	checkoutID := getStringField(event.Data, "id")
	checkoutStatus := getStringField(event.Data, "status")
	metadata := getMetadata(event.Data)
	ctx = injectTenantContext(ctx, metadata)

	promptID := metadata["prompt_id"]
	if promptID == "" {
		promptID = checkoutID
	}

	status := commonv1.STATUS_IN_PROCESS
	state := commonv1.STATE_ACTIVE
	switch checkoutStatus {
	case "succeeded", "confirmed":
		status = commonv1.STATUS_SUCCESSFUL
	case "failed", "expired":
		status = commonv1.STATUS_FAILED
		state = commonv1.STATE_INACTIVE
	}

	extras := data.JSONMap{
		"polar_event_type": event.Type,
		"checkout_id":      checkoutID,
		"checkout_status":  checkoutStatus,
		"entity_type":      "prompt",
	}

	statusReq := &commonv1.StatusUpdateRequest{
		Id:         promptID,
		State:      state,
		Status:     status,
		ExternalId: checkoutID,
		Extras:     extras.ToProtoStruct(),
	}

	if _, err := s.paymentCli.StatusUpdate(ctx, connect.NewRequest(statusReq)); err != nil {
		logger.WithError(err).Error("could not update payment status")
		return err
	}
	return nil
}

func (s *PolarWebhookServer) handleSubscriptionCreated(
	ctx context.Context,
	logger *util.LogEntry,
	event *client.WebhookEvent,
) error {
	subID := getStringField(event.Data, "id")
	metadata := getMetadata(event.Data)
	ctx = injectTenantContext(ctx, metadata)

	promptID := metadata["prompt_id"]
	if promptID == "" {
		promptID = subID
	}

	extras := data.JSONMap{
		"polar_event_type": event.Type,
		"subscription_id":  subID,
		"entity_type":      "prompt",
	}

	statusReq := &commonv1.StatusUpdateRequest{
		Id:         promptID,
		State:      commonv1.STATE_ACTIVE,
		Status:     commonv1.STATUS_SUCCESSFUL,
		ExternalId: subID,
		Extras:     extras.ToProtoStruct(),
	}

	if _, err := s.paymentCli.StatusUpdate(ctx, connect.NewRequest(statusReq)); err != nil {
		logger.WithError(err).Error("could not update payment status")
		return err
	}
	return nil
}

func (s *PolarWebhookServer) handleSubscriptionUpdated(
	ctx context.Context,
	logger *util.LogEntry,
	event *client.WebhookEvent,
) error {
	subID := getStringField(event.Data, "id")
	subStatus := getStringField(event.Data, "status")
	metadata := getMetadata(event.Data)
	ctx = injectTenantContext(ctx, metadata)

	promptID := metadata["prompt_id"]
	if promptID == "" {
		promptID = subID
	}

	status := commonv1.STATUS_IN_PROCESS
	state := commonv1.STATE_ACTIVE
	switch subStatus {
	case "active":
		status = commonv1.STATUS_SUCCESSFUL
	case "canceled", "past_due", "unpaid":
		status = commonv1.STATUS_FAILED
		state = commonv1.STATE_INACTIVE
	}

	extras := data.JSONMap{
		"polar_event_type":    event.Type,
		"subscription_id":     subID,
		"subscription_status": subStatus,
		"entity_type":         "prompt",
	}

	statusReq := &commonv1.StatusUpdateRequest{
		Id:         promptID,
		State:      state,
		Status:     status,
		ExternalId: subID,
		Extras:     extras.ToProtoStruct(),
	}

	if _, err := s.paymentCli.StatusUpdate(ctx, connect.NewRequest(statusReq)); err != nil {
		logger.WithError(err).Error("could not update payment status")
		return err
	}
	return nil
}

func getStringField(data map[string]any, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return ""
}

func getMetadata(eventData map[string]any) map[string]string {
	metadata := make(map[string]string)
	if meta, metaOk := eventData["metadata"].(map[string]any); metaOk {
		for k, v := range meta {
			if str, strOk := v.(string); strOk {
				metadata[k] = str
			}
		}
	}
	return metadata
}

// injectTenantContext creates AuthenticationClaims from webhook metadata and injects into context.
func injectTenantContext(ctx context.Context, metadata map[string]string) context.Context {
	tenantID := metadata["tenant_id"]
	partitionID := metadata["partition_id"]
	if tenantID == "" && partitionID == "" {
		return ctx
	}
	claims := &security.AuthenticationClaims{
		TenantID:    tenantID,
		PartitionID: partitionID,
	}
	return claims.ClaimsToContext(ctx)
}
