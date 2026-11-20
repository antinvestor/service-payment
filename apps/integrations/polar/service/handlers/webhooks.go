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

package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	"connectrpc.com/connect/v2"
	collectionv1 "github.com/antinvestor/service-payments/apps/billing/gen/collection/v1"
	"github.com/antinvestor/service-payments/apps/billing/gen/collection/v1/collectionv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/polar/config"
	"github.com/antinvestor/service-payments/apps/integrations/polar/service/client"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/events"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/pitabwire/frame/v3/data"
	"github.com/pitabwire/frame/v3/security"
	"github.com/pitabwire/util"
)

const (
	maxWebhookBodyBytes = 1 << 20
	providerName        = "polar"

	// Checkout session statuses (https://docs.polar.sh/api-reference/checkouts).
	checkoutStatusSucceeded = "succeeded"
	checkoutStatusConfirmed = "confirmed"
	checkoutStatusFailed    = "failed"
	checkoutStatusExpired   = "expired"

	// Order billing reasons.
	billingReasonSubscriptionCreate = "subscription_create"
	billingReasonSubscriptionCycle  = "subscription_cycle"
	billingReasonSubscriptionUpdate = "subscription_update"
)

// PolarWebhookServer handles Polar.sh webhook events.
//
// Checkout events settle the payment prompt that opened the checkout;
// subscription and order events are mirrored into billing, which owns the
// subscription state and the lifecycle fan-out to product systems.
type PolarWebhookServer struct {
	paymentCli    paymentv1connect.PaymentServiceClient
	billingCli    collectionv1connect.CollectionServiceClient
	polarCli      client.PolarClient
	webhookSecret string
	metrics       *integrationobs.Metrics
}

// NewPolarWebhookServer creates a new webhook server. billingCli may be nil:
// subscription and order events are then acknowledged without a mirror.
func NewPolarWebhookServer(
	paymentCli paymentv1connect.PaymentServiceClient,
	billingCli collectionv1connect.CollectionServiceClient,
	polarCli client.PolarClient,
	cfg *config.PolarConfig,
) *PolarWebhookServer {
	return &PolarWebhookServer{
		paymentCli:    paymentCli,
		billingCli:    billingCli,
		polarCli:      polarCli,
		webhookSecret: cfg.WebhookSecret,
		metrics:       integrationobs.NewMetrics(providerName),
	}
}

// NewRouterV1 creates the HTTP routes for Polar webhooks.
func (s *PolarWebhookServer) NewRouterV1() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/polar", s.HandleWebhook)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	return mux
}

// HandleWebhook processes Polar webhook events.
func (s *PolarWebhookServer) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.metrics.WebhookReceived(ctx, providerName)
	logger := util.Log(ctx).WithField("type", "polar.webhook")
	defer logger.Release()

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxWebhookBodyBytes+1))
	if err != nil || len(body) > maxWebhookBodyBytes {
		logger.Error("failed to read webhook body or body too large")
		s.metrics.WebhookRejected(ctx, providerName, "decode_error")
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
		s.metrics.WebhookRejected(ctx, providerName, "verification_failed")
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}
	if event.Timestamp == "" {
		event.Timestamp = headers["webhook-timestamp"]
	}

	logger = logger.WithField("event_type", event.Type).WithField("webhook_id", event.ID)
	logger.Debug("processing Polar webhook event")

	var handleErr error
	switch {
	case event.Type == "checkout.created":
		handleErr = s.handleCheckoutCreated(ctx, event)
	case event.Type == "checkout.updated":
		handleErr = s.handleCheckoutUpdated(ctx, event)
	case strings.HasPrefix(event.Type, "subscription."):
		handleErr = s.handleSubscriptionEvent(ctx, event)
	case strings.HasPrefix(event.Type, "order."):
		handleErr = s.handleOrderEvent(ctx, event)
	default:
		logger.Debug("unhandled event type, acknowledged")
	}

	if handleErr != nil {
		logger.WithError(handleErr).Error("webhook not applied; asking Polar to retry")
		s.metrics.WebhookRejected(ctx, providerName, "status_update_error")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ─── checkout events → payment prompt ────────────────────────────────────────

func (s *PolarWebhookServer) handleCheckoutCreated(ctx context.Context, event *client.WebhookEvent) error {
	checkoutID := getStringField(event.Data, "id")
	metadata := getMetadata(event.Data)
	ctx = injectTenantContext(ctx, metadata)
	promptID := metadata["prompt_id"]
	if promptID == "" {
		// Not one of ours (opened in the dashboard): nothing to settle.
		return nil
	}
	return s.updatePrompt(ctx, promptID, checkoutID, commonv1.STATUS_IN_PROCESS, data.JSONMap{
		"polar_event_type":          event.Type,
		"checkout_id":               checkoutID,
		"checkout_status":           getStringField(event.Data, "status"),
		collection.ExtraCheckoutURL: getStringField(event.Data, "url"),
		collection.ExtraNextAction:  collection.NextActionRedirectURL,
	})
}

func (s *PolarWebhookServer) handleCheckoutUpdated(ctx context.Context, event *client.WebhookEvent) error {
	checkoutID := getStringField(event.Data, "id")
	checkoutStatus := getStringField(event.Data, "status")
	metadata := getMetadata(event.Data)
	ctx = injectTenantContext(ctx, metadata)
	promptID := metadata["prompt_id"]
	if promptID == "" {
		return nil
	}

	status := commonv1.STATUS_IN_PROCESS
	switch checkoutStatus {
	case checkoutStatusSucceeded:
		status = commonv1.STATUS_SUCCESSFUL
	case checkoutStatusFailed, checkoutStatusExpired:
		status = commonv1.STATUS_FAILED
	case checkoutStatusConfirmed:
		// Payment submitted, Polar still processing: not yet success.
		status = commonv1.STATUS_IN_PROCESS
	}
	extras := data.JSONMap{
		"polar_event_type": event.Type,
		"checkout_id":      checkoutID,
		"checkout_status":  checkoutStatus,
	}
	if status == commonv1.STATUS_SUCCESSFUL {
		// Checkout reports what it charged; the session verifies it.
		if amt, cur, ok := checkoutAmount(event.Data); ok {
			extras["Amount"] = collection.DecimalString(amt)
			extras["Currency"] = cur
		}
		if subID := getStringField(event.Data, "subscription_id"); subID != "" {
			extras[collection.ExtraExternalSubscriptionID] = subID
		}
		if custID := getStringField(event.Data, "customer_id"); custID != "" {
			extras[collection.ExtraCustomerID] = custID
		}
	}
	if status == commonv1.STATUS_FAILED {
		extras[events.ExtraError] = "checkout " + checkoutStatus
	}
	return s.updatePrompt(ctx, promptID, checkoutID, status, extras)
}

func (s *PolarWebhookServer) updatePrompt(
	ctx context.Context,
	promptID, externalID string,
	status commonv1.STATUS,
	extras data.JSONMap,
) error {
	extras[events.ExtraEntityType] = events.EntityTypePrompt
	_, err := s.paymentCli.StatusUpdate(ctx, &commonv1.StatusUpdateRequest{
		Id:         promptID,
		State:      events.StateFor(status),
		Status:     status,
		ExternalId: externalID,
		Extras:     extras.ToProtoStruct(),
	})
	if err != nil {
		return err
	}
	return nil
}

// ─── subscription / order events → billing ───────────────────────────────────

// handleSubscriptionEvent mirrors subscription.{created,updated,active,
// canceled,uncanceled,revoked,past_due} into billing. The payment prompt
// that opened the subscription is also marked successful on activation, so
// a checkout session completes even when checkout.updated was missed.
func (s *PolarWebhookServer) handleSubscriptionEvent(ctx context.Context, event *client.WebhookEvent) error {
	metadata := getMetadata(event.Data)
	ctx = injectTenantContext(ctx, metadata)
	subID := getStringField(event.Data, "id")
	status := normalizeSubscriptionStatus(getStringField(event.Data, "status"))
	cancelAtPeriodEnd := getBoolField(event.Data, "cancel_at_period_end")

	req := &collectionv1.SyncProviderSubscriptionRequest{
		Provider:               providerName,
		Event:                  collection.SubscriptionEventUpdated,
		ExternalSubscriptionId: subID,
		ExternalCustomerId:     getStringField(event.Data, "customer_id"),
		ExternalProductId:      getStringField(event.Data, "product_id"),
		Status:                 status,
		CancelAtPeriodEnd:      cancelAtPeriodEnd,
		CurrentPeriodStart:     getStringField(event.Data, "current_period_start"),
		CurrentPeriodEnd:       getStringField(event.Data, "current_period_end"),
		SubscriptionId:         metadata[collection.ExtraSubscriptionID],
		ProfileId:              metadata["profile_id"],
		PlanId:                 metadata["plan_id"],
		OccurredAt:             firstNonEmpty(getStringField(event.Data, "modified_at"), event.Timestamp),
		Metadata:               metadata,
	}
	if err := s.syncBilling(ctx, req); err != nil {
		return err
	}

	// Activation also settles the prompt that opened it (idempotent).
	if promptID := metadata["prompt_id"]; promptID != "" &&
		(status == collection.SubscriptionStatusActive || status == collection.SubscriptionStatusTrialing) {
		return s.updatePrompt(ctx, promptID, subID, commonv1.STATUS_SUCCESSFUL, data.JSONMap{
			"polar_event_type":                     event.Type,
			"subscription_status":                  status,
			collection.ExtraExternalSubscriptionID: subID,
			collection.ExtraCustomerID:             getStringField(event.Data, "customer_id"),
			"current_period_end":                   getStringField(event.Data, "current_period_end"),
		})
	}
	return nil
}

// handleOrderEvent reports paid subscription orders (signup and renewals)
// to billing as billed events; one-time orders only settle their prompt.
func (s *PolarWebhookServer) handleOrderEvent(ctx context.Context, event *client.WebhookEvent) error {
	metadata := getMetadata(event.Data)
	ctx = injectTenantContext(ctx, metadata)
	orderID := getStringField(event.Data, "id")
	subID := getStringField(event.Data, "subscription_id")
	paid := getBoolField(event.Data, "paid") || event.Type == "order.paid" ||
		getStringField(event.Data, "status") == "paid"
	if !paid {
		return nil
	}
	amount, currency, hasAmount := orderAmount(event.Data)

	if subID != "" {
		reason := getStringField(event.Data, "billing_reason")
		if reason == "" || reason == billingReasonSubscriptionCreate ||
			reason == billingReasonSubscriptionCycle || reason == billingReasonSubscriptionUpdate {
			req := &collectionv1.SyncProviderSubscriptionRequest{
				Provider:               providerName,
				Event:                  collection.SubscriptionEventBilled,
				ExternalSubscriptionId: subID,
				ExternalCustomerId:     getStringField(event.Data, "customer_id"),
				ExternalProductId:      getStringField(event.Data, "product_id"),
				ExternalOrderId:        orderID,
				SubscriptionId:         metadata[collection.ExtraSubscriptionID],
				OccurredAt:             firstNonEmpty(getStringField(event.Data, "created_at"), event.Timestamp),
				Metadata:               metadata,
			}
			if nested, ok := event.Data["subscription"].(map[string]any); ok {
				req.CurrentPeriodStart = getStringField(nested, "current_period_start")
				req.CurrentPeriodEnd = getStringField(nested, "current_period_end")
			}
			if hasAmount {
				req.Amount = amount
			}
			if err := s.syncBilling(ctx, req); err != nil {
				return err
			}
		}
	}

	// A one-time purchase order settles its prompt directly.
	if promptID := metadata["prompt_id"]; promptID != "" && subID == "" {
		extras := data.JSONMap{"polar_event_type": event.Type, "order_id": orderID}
		if hasAmount {
			extras["Amount"] = collection.DecimalString(amount)
			extras["Currency"] = currency
		}
		return s.updatePrompt(ctx, promptID, orderID, commonv1.STATUS_SUCCESSFUL, extras)
	}
	return nil
}

// syncBilling calls billing; an unknown subscription is logged and dropped
// (Polar would otherwise retry forever), anything else is retried.
func (s *PolarWebhookServer) syncBilling(ctx context.Context, req *collectionv1.SyncProviderSubscriptionRequest) error {
	if s.billingCli == nil {
		util.Log(ctx).WithField("event", req.GetEvent()).Debug("billing client not configured; subscription event acknowledged")
		return nil
	}
	resp, err := s.billingCli.SyncProviderSubscription(ctx, req)
	if err != nil {
		var connectErr *connect.Error
		if errors.As(err, &connectErr) && connectErr.Code() == connect.CodeNotFound {
			util.Log(ctx).WithError(err).
				WithField("external_subscription_id", req.GetExternalSubscriptionId()).
				Warn("no billing subscription for Polar subscription; event dropped")
			s.metrics.WebhookRejected(ctx, providerName, "unknown_payment")
			return nil
		}
		return err
	}
	util.Log(ctx).WithField("subscription_id", resp.GetSubscriptionId()).
		WithField("state", resp.GetSubscriptionState()).
		WithField("applied", resp.GetApplied()).
		Debug("billing subscription mirrored")
	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// normalizeSubscriptionStatus maps Polar's status vocabulary onto the
// portable set in pkg/collection (they coincide today).
func normalizeSubscriptionStatus(status string) string {
	s := strings.ToLower(strings.TrimSpace(status))
	switch s {
	case "cancelled":
		return collection.SubscriptionStatusCanceled
	default:
		return s
	}
}

func checkoutAmount(d map[string]any) (*commonv1.Money, string, bool) {
	currency := strings.ToUpper(getStringField(d, "currency"))
	for _, key := range []string{"total_amount", "amount"} {
		if minor, ok := getIntField(d, key); ok && currency != "" {
			return collection.FromMinorUnits(minor, currency), currency, true
		}
	}
	return nil, "", false
}

func orderAmount(d map[string]any) (*commonv1.Money, string, bool) {
	currency := strings.ToUpper(getStringField(d, "currency"))
	for _, key := range []string{"total_amount", "amount", "net_amount"} {
		if minor, ok := getIntField(d, key); ok && currency != "" {
			return collection.FromMinorUnits(minor, currency), currency, true
		}
	}
	return nil, "", false
}

func getStringField(d map[string]any, key string) string {
	if v, ok := d[key].(string); ok {
		return v
	}
	return ""
}

func getBoolField(d map[string]any, key string) bool {
	v, _ := d[key].(bool)
	return v
}

func getIntField(d map[string]any, key string) (int64, bool) {
	switch v := d[key].(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case int:
		return int64(v), true
	default:
		return 0, false
	}
}

func getMetadata(eventData map[string]any) map[string]string {
	metadata := make(map[string]string)
	if meta, metaOk := eventData["metadata"].(map[string]any); metaOk {
		for k, v := range meta {
			switch t := v.(type) {
			case string:
				metadata[k] = t
			case float64:
				metadata[k] = formatFloat(t)
			case bool:
				if t {
					metadata[k] = "true"
				} else {
					metadata[k] = "false"
				}
			}
		}
	}
	return metadata
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
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
