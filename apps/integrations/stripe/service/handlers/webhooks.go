package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/stripe/config"
	"github.com/antinvestor/service-payments/apps/integrations/stripe/service/client"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/security"
	"github.com/pitabwire/util"
)

// StripeWebhookServer handles Stripe webhook events.
type StripeWebhookServer struct {
	paymentCli    paymentv1connect.PaymentServiceClient
	stripeCli     client.StripeClient
	webhookSecret string
}

// NewStripeWebhookServer creates a new webhook server.
func NewStripeWebhookServer(
	paymentCli paymentv1connect.PaymentServiceClient,
	stripeCli client.StripeClient,
	cfg *config.StripeConfig,
) *StripeWebhookServer {
	return &StripeWebhookServer{
		paymentCli:    paymentCli,
		stripeCli:     stripeCli,
		webhookSecret: cfg.WebhookSecret,
	}
}

// NewRouterV1 creates the HTTP routes for Stripe webhooks.
func (s *StripeWebhookServer) NewRouterV1() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/stripe", s.HandleWebhook)
	return mux
}

// HandleWebhook processes all Stripe webhook events.
func (s *StripeWebhookServer) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := util.Log(ctx).WithField("type", "stripe.webhook")
	defer logger.Release()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.WithError(err).Error("failed to read webhook body")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	signature := r.Header.Get("Stripe-Signature")
	event, err := s.stripeCli.VerifyWebhookSignature(body, signature, s.webhookSecret)
	if err != nil {
		logger.WithError(err).Error("webhook signature verification failed")
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	logger = logger.WithFields(map[string]any{"event_type": event.Type, "event_id": event.ID})
	logger.Debug("processing Stripe webhook event")

	var statusErr error
	switch event.Type {
	case "payment_intent.succeeded":
		statusErr = s.handlePaymentIntentSucceeded(ctx, logger, event)
	case "payment_intent.payment_failed":
		statusErr = s.handlePaymentIntentFailed(ctx, logger, event)
	case "payout.paid":
		statusErr = s.handlePayoutPaid(ctx, logger, event)
	case "payout.failed":
		statusErr = s.handlePayoutFailed(ctx, logger, event)
	default:
		logger.Debug("unhandled event type, skipping")
	}

	if statusErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *StripeWebhookServer) handlePaymentIntentSucceeded(
	ctx context.Context,
	logger *util.LogEntry,
	event *client.WebhookEvent,
) error {
	objID, metadata := extractObjectData(event)
	ctx = injectTenantContext(ctx, metadata)

	paymentID := metadata["prompt_id"]
	if paymentID == "" {
		paymentID = objID
	}

	extras := data.JSONMap{
		"stripe_event_id":   event.ID,
		"stripe_event_type": event.Type,
		"stripe_object_id":  objID,
		"entity_type":       "prompt",
	}

	statusReq := &commonv1.StatusUpdateRequest{
		Id:         paymentID,
		State:      commonv1.STATE_ACTIVE,
		Status:     commonv1.STATUS_SUCCESSFUL,
		ExternalId: objID,
		Extras:     extras.ToProtoStruct(),
	}

	if _, err := s.paymentCli.StatusUpdate(ctx, connect.NewRequest(statusReq)); err != nil {
		logger.WithError(err).Error("could not update payment status")
		return err
	}
	return nil
}

func (s *StripeWebhookServer) handlePaymentIntentFailed(
	ctx context.Context,
	logger *util.LogEntry,
	event *client.WebhookEvent,
) error {
	objID, metadata := extractObjectData(event)
	ctx = injectTenantContext(ctx, metadata)

	paymentID := metadata["prompt_id"]
	if paymentID == "" {
		paymentID = objID
	}

	extras := data.JSONMap{
		"stripe_event_id":   event.ID,
		"stripe_event_type": event.Type,
		"stripe_object_id":  objID,
		"entity_type":       "prompt",
	}

	statusReq := &commonv1.StatusUpdateRequest{
		Id:         paymentID,
		State:      commonv1.STATE_INACTIVE,
		Status:     commonv1.STATUS_FAILED,
		ExternalId: objID,
		Extras:     extras.ToProtoStruct(),
	}

	if _, err := s.paymentCli.StatusUpdate(ctx, connect.NewRequest(statusReq)); err != nil {
		logger.WithError(err).Error("could not update payment status")
		return err
	}
	return nil
}

func (s *StripeWebhookServer) handlePayoutPaid(
	ctx context.Context,
	logger *util.LogEntry,
	event *client.WebhookEvent,
) error {
	objID, metadata := extractObjectData(event)
	ctx = injectTenantContext(ctx, metadata)

	paymentID := metadata["payment_id"]
	if paymentID == "" {
		paymentID = objID
	}

	extras := data.JSONMap{
		"stripe_event_id":   event.ID,
		"stripe_event_type": event.Type,
		"stripe_object_id":  objID,
		"entity_type":       "payment",
	}

	statusReq := &commonv1.StatusUpdateRequest{
		Id:         paymentID,
		State:      commonv1.STATE_ACTIVE,
		Status:     commonv1.STATUS_SUCCESSFUL,
		ExternalId: objID,
		Extras:     extras.ToProtoStruct(),
	}

	if _, err := s.paymentCli.StatusUpdate(ctx, connect.NewRequest(statusReq)); err != nil {
		logger.WithError(err).Error("could not update payment status")
		return err
	}
	return nil
}

func (s *StripeWebhookServer) handlePayoutFailed(
	ctx context.Context,
	logger *util.LogEntry,
	event *client.WebhookEvent,
) error {
	objID, metadata := extractObjectData(event)
	ctx = injectTenantContext(ctx, metadata)

	paymentID := metadata["payment_id"]
	if paymentID == "" {
		paymentID = objID
	}

	extras := data.JSONMap{
		"stripe_event_id":   event.ID,
		"stripe_event_type": event.Type,
		"stripe_object_id":  objID,
		"entity_type":       "payment",
	}

	statusReq := &commonv1.StatusUpdateRequest{
		Id:         paymentID,
		State:      commonv1.STATE_INACTIVE,
		Status:     commonv1.STATUS_FAILED,
		ExternalId: objID,
		Extras:     extras.ToProtoStruct(),
	}

	if _, err := s.paymentCli.StatusUpdate(ctx, connect.NewRequest(statusReq)); err != nil {
		logger.WithError(err).Error("could not update payment status")
		return err
	}
	return nil
}

// extractObjectData extracts the object ID and metadata from a webhook event's raw data.
func extractObjectData(event *client.WebhookEvent) (string, map[string]string) {
	metadata := make(map[string]string)
	objID := ""

	rawData, rawOk := event.Data["raw"].(string)
	if !rawOk {
		return objID, metadata
	}

	var obj map[string]any
	if err := json.Unmarshal([]byte(rawData), &obj); err != nil {
		return objID, metadata
	}

	if id, idOk := obj["id"].(string); idOk {
		objID = id
	}

	if meta, metaOk := obj["metadata"].(map[string]any); metaOk {
		for k, v := range meta {
			if str, strOk := v.(string); strOk {
				metadata[k] = str
			}
		}
	}

	return objID, metadata
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
