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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	"connectrpc.com/connect/v2"
	collectionv1 "github.com/antinvestor/service-payments/apps/billing/gen/collection/v1"
	"github.com/antinvestor/service-payments/apps/billing/gen/collection/v1/collectionv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/stripe/config"
	"github.com/antinvestor/service-payments/apps/integrations/stripe/service/client"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/events"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/pitabwire/frame/v3/data"
	"github.com/pitabwire/frame/v3/security"
	"github.com/pitabwire/util"
	"github.com/stripe/stripe-go/v82"
)

const (
	maxWebhookBodyBytes = 1 << 20
	providerName        = "stripe"

	extraEventID   = "stripe_event_id"
	extraEventType = "stripe_event_type"

	// Prompt status extras set by the prompt worker (kept in sync by name).
	extraCheckoutSessionID = "checkout_session_id"
	extraPaymentIntentID   = "payment_intent_id"
	extraPayoutID          = "payout_id"

	// Stripe event types handled here.
	eventCheckoutCompleted      = "checkout.session.completed"
	eventCheckoutAsyncSucceeded = "checkout.session.async_payment_succeeded"
	eventCheckoutAsyncFailed    = "checkout.session.async_payment_failed"
	eventCheckoutExpired        = "checkout.session.expired"
	eventPaymentIntentSucceeded = "payment_intent.succeeded"
	eventPaymentIntentFailed    = "payment_intent.payment_failed"
	eventPayoutPaid             = "payout.paid"
	eventPayoutFailed           = "payout.failed"
	eventInvoicePaid            = "invoice.paid"
	eventInvoicePaymentFailed   = "invoice.payment_failed"
	subscriptionEventPrefix     = "customer.subscription."
)

// StripeWebhookServer handles Stripe webhook events.
//
// Checkout and payment intent events settle the payment prompt that opened
// them; subscription and invoice events are mirrored into billing, which
// owns the subscription state and the lifecycle fan-out to product systems.
type StripeWebhookServer struct {
	paymentCli    paymentv1connect.PaymentServiceClient
	billingCli    collectionv1connect.CollectionServiceClient
	stripeCli     client.StripeClient
	webhookSecret string
	metrics       *integrationobs.Metrics
}

// NewStripeWebhookServer creates a new webhook server. billingCli may be nil:
// subscription and invoice events are then acknowledged without a mirror.
func NewStripeWebhookServer(
	paymentCli paymentv1connect.PaymentServiceClient,
	billingCli collectionv1connect.CollectionServiceClient,
	stripeCli client.StripeClient,
	cfg *config.StripeConfig,
) *StripeWebhookServer {
	return &StripeWebhookServer{
		paymentCli:    paymentCli,
		billingCli:    billingCli,
		stripeCli:     stripeCli,
		webhookSecret: cfg.WebhookSecret,
		metrics:       integrationobs.NewMetrics(providerName),
	}
}

// NewRouterV1 creates the HTTP routes for Stripe webhooks.
func (s *StripeWebhookServer) NewRouterV1() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/stripe", s.HandleWebhook)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	return mux
}

// HandleWebhook processes Stripe webhook events.
func (s *StripeWebhookServer) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.metrics.WebhookReceived(ctx, providerName)
	logger := util.Log(ctx).WithField("type", "stripe.webhook")
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

	event, err := s.stripeCli.VerifyWebhookSignature(body, r.Header.Get("Stripe-Signature"), s.webhookSecret)
	if err != nil {
		logger.WithError(err).Error("webhook signature verification failed")
		s.metrics.WebhookRejected(ctx, providerName, "verification_failed")
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	eventType := string(event.Type)
	logger = logger.WithField("event_type", eventType).WithField("event_id", event.ID)
	logger.Debug("processing Stripe webhook event")

	var handleErr error
	switch {
	case eventType == eventCheckoutCompleted, eventType == eventCheckoutAsyncSucceeded,
		eventType == eventCheckoutAsyncFailed, eventType == eventCheckoutExpired:
		handleErr = s.handleCheckoutSession(ctx, event)
	case eventType == eventPaymentIntentSucceeded, eventType == eventPaymentIntentFailed:
		handleErr = s.handlePaymentIntent(ctx, event)
	case eventType == eventPayoutPaid, eventType == eventPayoutFailed:
		handleErr = s.handlePayout(ctx, event)
	case strings.HasPrefix(eventType, subscriptionEventPrefix):
		handleErr = s.handleSubscription(ctx, event)
	case eventType == eventInvoicePaid, eventType == eventInvoicePaymentFailed:
		handleErr = s.handleInvoice(ctx, event)
	default:
		logger.Debug("unhandled event type, acknowledged")
	}

	if handleErr != nil {
		logger.WithError(handleErr).Error("webhook not applied; asking Stripe to retry")
		s.metrics.WebhookRejected(ctx, providerName, "status_update_error")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ─── checkout / payment intent / payout → payment service ────────────────────

func (s *StripeWebhookServer) handleCheckoutSession(ctx context.Context, event *stripe.Event) error {
	var sess stripe.CheckoutSession
	if err := decodeObject(event, &sess); err != nil {
		return err
	}
	ctx = injectTenantContext(ctx, sess.Metadata)
	promptID := sess.Metadata["prompt_id"]
	if promptID == "" {
		// Not one of ours (opened in the dashboard): nothing to settle.
		return nil
	}

	eventType := string(event.Type)
	extras := data.JSONMap{
		extraEventID:           event.ID,
		extraEventType:         eventType,
		extraCheckoutSessionID: sess.ID,
		"checkout_status":      string(sess.Status),
		"payment_status":       string(sess.PaymentStatus),
	}
	status := commonv1.STATUS_IN_PROCESS
	switch eventType {
	case eventCheckoutCompleted, eventCheckoutAsyncSucceeded:
		status = commonv1.STATUS_SUCCESSFUL
		if eventType == eventCheckoutCompleted && sess.PaymentStatus == stripe.CheckoutSessionPaymentStatusUnpaid {
			// Delayed-notification method (bank debit, …): the payer is
			// done, the money is not; async_payment_succeeded settles it.
			status = commonv1.STATUS_IN_PROCESS
		}
	case eventCheckoutAsyncFailed, eventCheckoutExpired:
		status = commonv1.STATUS_FAILED
		extras[events.ExtraError] = "checkout " + strings.TrimPrefix(eventType, "checkout.session.")
	}
	if status == commonv1.STATUS_SUCCESSFUL {
		// Stripe reports what it charged; the session verifies it.
		if sess.Currency != "" && sess.AmountTotal > 0 {
			extras["Amount"] = collection.DecimalString(collection.FromMinorUnits(sess.AmountTotal, string(sess.Currency)))
			extras["Currency"] = strings.ToUpper(string(sess.Currency))
		}
		if sess.Subscription != nil && sess.Subscription.ID != "" {
			extras[collection.ExtraExternalSubscriptionID] = sess.Subscription.ID
		}
		if sess.Customer != nil && sess.Customer.ID != "" {
			extras[collection.ExtraCustomerID] = sess.Customer.ID
		}
		if sess.PaymentIntent != nil && sess.PaymentIntent.ID != "" {
			extras[extraPaymentIntentID] = sess.PaymentIntent.ID
		}
	}
	return s.updateStatus(ctx, events.EntityTypePrompt, promptID, sess.ID, status, extras)
}

func (s *StripeWebhookServer) handlePaymentIntent(ctx context.Context, event *stripe.Event) error {
	var pi stripe.PaymentIntent
	if err := decodeObject(event, &pi); err != nil {
		return err
	}
	ctx = injectTenantContext(ctx, pi.Metadata)
	promptID := pi.Metadata["prompt_id"]
	if promptID == "" {
		return nil
	}
	extras := data.JSONMap{
		extraEventID:         event.ID,
		extraEventType:       string(event.Type),
		extraPaymentIntentID: pi.ID,
		"stripe_status":      string(pi.Status),
	}
	status := commonv1.STATUS_SUCCESSFUL
	if string(event.Type) == eventPaymentIntentFailed {
		status = commonv1.STATUS_FAILED
		reason := "payment failed"
		if pi.LastPaymentError != nil && pi.LastPaymentError.Msg != "" {
			reason = pi.LastPaymentError.Msg
		}
		extras[events.ExtraError] = reason
	} else if pi.Currency != "" {
		amount := pi.AmountReceived
		if amount == 0 {
			amount = pi.Amount
		}
		extras["Amount"] = collection.DecimalString(collection.FromMinorUnits(amount, string(pi.Currency)))
		extras["Currency"] = strings.ToUpper(string(pi.Currency))
	}
	return s.updateStatus(ctx, events.EntityTypePrompt, promptID, pi.ID, status, extras)
}

func (s *StripeWebhookServer) handlePayout(ctx context.Context, event *stripe.Event) error {
	var po stripe.Payout
	if err := decodeObject(event, &po); err != nil {
		return err
	}
	ctx = injectTenantContext(ctx, po.Metadata)
	paymentID := po.Metadata["payment_id"]
	if paymentID == "" {
		return nil
	}
	extras := data.JSONMap{
		extraEventID:    event.ID,
		extraEventType:  string(event.Type),
		extraPayoutID:   po.ID,
		"stripe_status": string(po.Status),
	}
	status := commonv1.STATUS_SUCCESSFUL
	if string(event.Type) == eventPayoutFailed {
		status = commonv1.STATUS_FAILED
		extras[events.ExtraError] = firstNonEmpty(po.FailureMessage, string(po.FailureCode), "payout failed")
	} else if po.Currency != "" {
		extras["Amount"] = collection.DecimalString(collection.FromMinorUnits(po.Amount, string(po.Currency)))
		extras["Currency"] = strings.ToUpper(string(po.Currency))
	}
	return s.updateStatus(ctx, events.EntityTypePayment, paymentID, po.ID, status, extras)
}

func (s *StripeWebhookServer) updateStatus(
	ctx context.Context,
	entityType, id, externalID string,
	status commonv1.STATUS,
	extras data.JSONMap,
) error {
	extras[events.ExtraEntityType] = entityType
	_, err := s.paymentCli.StatusUpdate(ctx, &commonv1.StatusUpdateRequest{
		Id:         id,
		State:      events.StateFor(status),
		Status:     status,
		ExternalId: externalID,
		Extras:     extras.ToProtoStruct(),
	})
	if err != nil {
		return fmt.Errorf("update %s %s status: %w", entityType, id, err)
	}
	return nil
}

// ─── subscription / invoice events → billing ─────────────────────────────────

// handleSubscription mirrors customer.subscription.{created,updated,deleted,
// paused,resumed,…} into billing. The payment prompt that opened the
// subscription is also marked successful on activation, so a checkout
// completes even when checkout.session.completed was missed.
func (s *StripeWebhookServer) handleSubscription(ctx context.Context, event *stripe.Event) error {
	var sub stripe.Subscription
	if err := decodeObject(event, &sub); err != nil {
		return err
	}
	ctx = injectTenantContext(ctx, sub.Metadata)
	status := normalizeSubscriptionStatus(string(sub.Status))
	periodStart, periodEnd := subscriptionPeriod(&sub, event.Data.Object)

	req := &collectionv1.SyncProviderSubscriptionRequest{
		Provider:               providerName,
		Event:                  collection.SubscriptionEventUpdated,
		ExternalSubscriptionId: sub.ID,
		ExternalCustomerId:     customerID(sub.Customer),
		ExternalProductId:      subscriptionPriceID(&sub),
		Status:                 status,
		CancelAtPeriodEnd:      sub.CancelAtPeriodEnd,
		CurrentPeriodStart:     periodStart,
		CurrentPeriodEnd:       periodEnd,
		SubscriptionId:         sub.Metadata[collection.ExtraSubscriptionID],
		ProfileId:              sub.Metadata["profile_id"],
		PlanId:                 sub.Metadata["plan_id"],
		OccurredAt:             unixRFC3339(event.Created),
		Metadata:               sub.Metadata,
	}
	if err := s.syncBilling(ctx, req); err != nil {
		return err
	}

	// Activation also settles the prompt that opened it (idempotent).
	if promptID := sub.Metadata["prompt_id"]; promptID != "" &&
		(status == collection.SubscriptionStatusActive || status == collection.SubscriptionStatusTrialing) {
		return s.updateStatus(ctx, events.EntityTypePrompt, promptID, sub.ID, commonv1.STATUS_SUCCESSFUL, data.JSONMap{
			extraEventID:                           event.ID,
			extraEventType:                         string(event.Type),
			"subscription_status":                  status,
			collection.ExtraExternalSubscriptionID: sub.ID,
			collection.ExtraCustomerID:             customerID(sub.Customer),
			"current_period_end":                   periodEnd,
		})
	}
	return nil
}

// handleInvoice reports subscription invoices to billing: paid ones as
// billed events (signup and renewals), failed ones as payment failures.
// Invoices without a subscription are not billing's concern.
func (s *StripeWebhookServer) handleInvoice(ctx context.Context, event *stripe.Event) error {
	var inv stripe.Invoice
	if err := decodeObject(event, &inv); err != nil {
		return err
	}
	subID, subMetadata := invoiceSubscription(&inv, event.Data.Object)
	if subID == "" {
		return nil
	}
	metadata := mergeMetadata(subMetadata, inv.Metadata)
	ctx = injectTenantContext(ctx, metadata)

	req := &collectionv1.SyncProviderSubscriptionRequest{
		Provider:               providerName,
		Event:                  collection.SubscriptionEventBilled,
		ExternalSubscriptionId: subID,
		ExternalCustomerId:     customerID(inv.Customer),
		ExternalOrderId:        inv.ID,
		SubscriptionId:         metadata[collection.ExtraSubscriptionID],
		ProfileId:              metadata["profile_id"],
		PlanId:                 metadata["plan_id"],
		CurrentPeriodStart:     unixRFC3339(inv.PeriodStart),
		CurrentPeriodEnd:       unixRFC3339(inv.PeriodEnd),
		OccurredAt:             unixRFC3339(event.Created),
		Metadata:               metadata,
	}
	if string(event.Type) == eventInvoicePaymentFailed {
		req.Event = collection.SubscriptionEventPaymentFailed
	}
	if inv.Currency != "" {
		amount := inv.AmountPaid
		if req.GetEvent() == collection.SubscriptionEventPaymentFailed || amount == 0 {
			amount = inv.AmountDue
		}
		if amount > 0 {
			req.Amount = collection.FromMinorUnits(amount, string(inv.Currency))
		}
	}
	if line := firstInvoiceLine(&inv); line != nil {
		if line.Pricing != nil && line.Pricing.PriceDetails != nil {
			req.ExternalProductId = line.Pricing.PriceDetails.Price
		}
		if line.Period != nil && line.Period.Start > 0 && line.Period.End > 0 {
			req.CurrentPeriodStart = unixRFC3339(line.Period.Start)
			req.CurrentPeriodEnd = unixRFC3339(line.Period.End)
		}
	}
	return s.syncBilling(ctx, req)
}

// syncBilling calls billing; an unknown subscription is logged and dropped
// (Stripe would otherwise retry for days), anything else is retried.
func (s *StripeWebhookServer) syncBilling(ctx context.Context, req *collectionv1.SyncProviderSubscriptionRequest) error {
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
				Warn("no billing subscription for Stripe subscription; event dropped")
			s.metrics.WebhookRejected(ctx, providerName, "unknown_payment")
			return nil
		}
		return fmt.Errorf("sync billing subscription %s: %w", req.GetExternalSubscriptionId(), err)
	}
	util.Log(ctx).WithField("subscription_id", resp.GetSubscriptionId()).
		WithField("state", resp.GetSubscriptionState()).
		WithField("applied", resp.GetApplied()).
		Debug("billing subscription mirrored")
	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// decodeObject unmarshals the event's object into the typed stripe-go struct.
func decodeObject(event *stripe.Event, out any) error {
	if event.Data == nil || len(event.Data.Raw) == 0 {
		return fmt.Errorf("event %s carries no object", event.ID)
	}
	if err := json.Unmarshal(event.Data.Raw, out); err != nil {
		return fmt.Errorf("decode %s object: %w", event.Type, err)
	}
	if event.Data.Object == nil {
		// Events built in code (tests) skip stripe-go's EventData decoder.
		_ = json.Unmarshal(event.Data.Raw, &event.Data.Object)
	}
	return nil
}

// normalizeSubscriptionStatus maps Stripe's status vocabulary onto the
// portable set in pkg/collection (they coincide).
func normalizeSubscriptionStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

// subscriptionPeriod reads the current period. API versions from 2025-03-31
// put it on the subscription items; older ones on the subscription itself.
func subscriptionPeriod(sub *stripe.Subscription, raw map[string]any) (string, string) {
	if sub.Items != nil && len(sub.Items.Data) > 0 && sub.Items.Data[0] != nil {
		if item := sub.Items.Data[0]; item.CurrentPeriodEnd > 0 {
			return unixRFC3339(item.CurrentPeriodStart), unixRFC3339(item.CurrentPeriodEnd)
		}
	}
	return unixRFC3339(rawInt(raw, "current_period_start")), unixRFC3339(rawInt(raw, "current_period_end"))
}

func subscriptionPriceID(sub *stripe.Subscription) string {
	if sub.Items == nil || len(sub.Items.Data) == 0 || sub.Items.Data[0] == nil || sub.Items.Data[0].Price == nil {
		return ""
	}
	return sub.Items.Data[0].Price.ID
}

// invoiceSubscription returns the subscription that generated the invoice
// and the subscription metadata snapshot. API versions from 2025-03-31
// carry both under parent.subscription_details; older ones have a
// top-level subscription id.
func invoiceSubscription(inv *stripe.Invoice, raw map[string]any) (string, map[string]string) {
	if inv.Parent != nil && inv.Parent.SubscriptionDetails != nil {
		details := inv.Parent.SubscriptionDetails
		if details.Subscription != nil && details.Subscription.ID != "" {
			return details.Subscription.ID, details.Metadata
		}
	}
	switch v := raw["subscription"].(type) {
	case string:
		return v, nil
	case map[string]any:
		id, _ := v["id"].(string)
		return id, nil
	}
	if line := firstInvoiceLine(inv); line != nil && line.Subscription != nil {
		return line.Subscription.ID, nil
	}
	return "", nil
}

func firstInvoiceLine(inv *stripe.Invoice) *stripe.InvoiceLineItem {
	if inv.Lines == nil || len(inv.Lines.Data) == 0 {
		return nil
	}
	return inv.Lines.Data[0]
}

func customerID(c *stripe.Customer) string {
	if c == nil {
		return ""
	}
	return c.ID
}

func mergeMetadata(base, override map[string]string) map[string]string {
	out := make(map[string]string, len(base)+len(override))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		if v != "" {
			out[k] = v
		}
	}
	return out
}

func rawInt(raw map[string]any, key string) int64 {
	switch v := raw[key].(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case json.Number:
		n, _ := v.Int64()
		return n
	}
	return 0
}

func unixRFC3339(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return time.Unix(ts, 0).UTC().Format(time.RFC3339)
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
