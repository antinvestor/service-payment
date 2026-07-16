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
	"fmt"
	"io"
	"net/http"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/config"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/client"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/frame/v2/security"
	"github.com/pitabwire/util"
)

// FlutterwaveWebhookServer handles Flutterwave v4 webhooks.
// Docs: https://developer.flutterwave.com/docs/webhooks
type FlutterwaveWebhookServer struct {
	paymentCli    paymentv1connect.PaymentServiceClient
	fwCli         client.FlutterwaveClient
	webhookSecret string
	// defaultCreds used for charge re-verify after webhook.
	defaultCreds *client.Credentials
	metrics      *integrationobs.Metrics
}

// NewFlutterwaveWebhookServer constructs the HTTP webhook server.
func NewFlutterwaveWebhookServer(
	paymentCli paymentv1connect.PaymentServiceClient,
	fwCli client.FlutterwaveClient,
	cfg *config.FlutterwaveConfig,
) *FlutterwaveWebhookServer {
	creds := &client.Credentials{
		ClientID:      cfg.ClientID,
		ClientSecret:  cfg.ClientSecret,
		PublicKey:     firstNonEmpty(cfg.PublicKey, cfg.ClientID),
		SecretKey:     firstNonEmpty(cfg.SecretKey, cfg.ClientSecret),
		EncryptionKey: cfg.EncryptionKey,
		WebhookSecret: cfg.WebhookSecret,
		Environment:   cfg.Environment,
		OAuthTokenURL: cfg.OAuthTokenURL,
	}
	if strings.HasPrefix(creds.ClientID, "FLWPUBK_") {
		creds.PublicKey = creds.ClientID
	}
	if strings.HasPrefix(creds.ClientSecret, "FLWSECK_") {
		creds.SecretKey = creds.ClientSecret
	}
	if strings.EqualFold(cfg.Environment, "production") {
		creds.APIBaseURL = cfg.ProductionAPIBaseURL
	} else {
		creds.APIBaseURL = cfg.SandboxAPIBaseURL
	}
	return &FlutterwaveWebhookServer{
		paymentCli:    paymentCli,
		fwCli:         fwCli,
		webhookSecret: cfg.WebhookSecret,
		defaultCreds:  creds,
		metrics:       integrationobs.NewMetrics("flutterwave"),
	}
}

// NewRouterV1 registers webhook routes.
func (s *FlutterwaveWebhookServer) NewRouterV1() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /webhook/flutterwave", s.HandleWebhook)
	mux.HandleFunc("GET /webhook/flutterwave/return", s.HandleReturn)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

// HandleWebhook processes charge.completed, transfer.disburse, etc.
func (s *FlutterwaveWebhookServer) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.metrics.WebhookReceived(ctx, "flutterwave")
	logger := util.Log(ctx).WithField("type", "flutterwave.webhook")
	defer logger.Release()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.metrics.WebhookRejected(ctx, "flutterwave", "decode_error")
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}

	// v4: Flutterwave-Signature = base64(HMAC-SHA256(secret_hash, raw_body))
	// Also accept legacy Verif-Hash exact match for dashboards still on that mode.
	sig := r.Header.Get("Flutterwave-Signature")
	if s.webhookSecret != "" {
		ok := s.fwCli.VerifyWebhookSignature(body, sig, s.webhookSecret)
		if !ok {
			verif := r.Header.Get("Verif-Hash")
			if verif == "" || verif != s.webhookSecret {
				logger.Error("webhook signature verification failed")
				s.metrics.WebhookRejected(ctx, "flutterwave", "verification_failed")
				http.Error(w, "invalid signature", http.StatusUnauthorized)
				return
			}
		}
	}

	var event client.WebhookEvent
	if err = json.Unmarshal(body, &event); err != nil {
		s.metrics.WebhookRejected(ctx, "flutterwave", "decode_error")
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	eventType := event.EventType()
	logger = logger.WithField("event_type", eventType).WithField("webhook_id", event.WebhookID)
	logger.Debug("processing flutterwave webhook")

	ctx = injectTenantFromQuery(ctx, r)
	ctx = injectTenantFromMeta(ctx, event.Data)

	var processErr error
	switch strings.ToLower(eventType) {
	case "charge.completed":
		processErr = s.handleChargeCompleted(ctx, logger, &event)
	case "transfer.disburse", "transfer.completed":
		processErr = s.handleTransferDisburse(ctx, logger, &event)
	case "transfer.reversal":
		processErr = s.handleTransferDisburse(ctx, logger, &event) // map similarly
	case "order.authorization":
		logger.Debug("order.authorization acknowledged")
	default:
		logger.Debug("unhandled event type")
	}

	if processErr != nil {
		s.metrics.WebhookRejected(ctx, "flutterwave", "status_update_error")
		logger.WithError(processErr).Error("webhook processing error")
		http.Error(w, "processing error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// HandleReturn is the browser landing after redirect_url next_action.
func (s *FlutterwaveWebhookServer) HandleReturn(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	// Prefer charge id if present for best-effort reconcile.
	chargeID := q.Get("charge_id")
	if chargeID == "" {
		chargeID = q.Get("id")
	}
	if chargeID != "" && s.defaultCreds != nil &&
		(s.defaultCreds.ClientID != "" || s.defaultCreds.SecretKey != "" || s.defaultCreds.ClientSecret != "") {
		go s.reconcileCharge(context.WithoutCancel(r.Context()), chargeID)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<!doctype html><html><body style="font-family:system-ui;padding:2rem">
<h1>Payment received</h1>
<p>You can close this window. Your merchant will confirm shortly.</p>
</body></html>`))
}

func (s *FlutterwaveWebhookServer) reconcileCharge(ctx context.Context, chargeID string) {
	logger := util.Log(ctx).WithField("type", "flutterwave.return_reconcile")
	ch, err := s.fwCli.GetCharge(ctx, s.defaultCreds, chargeID)
	if err != nil {
		logger.WithError(err).Warn("return get charge failed")
		return
	}
	_ = s.applyCharge(ctx, logger, ch)
}

func (s *FlutterwaveWebhookServer) handleChargeCompleted(
	ctx context.Context,
	logger *util.LogEntry,
	event *client.WebhookEvent,
) error {
	chargeID := anyToString(event.Data["id"])
	// Re-verify with API when credentials available (best practice).
	if chargeID != "" && s.canVerify() {
		ch, err := s.fwCli.GetCharge(ctx, s.defaultCreds, chargeID)
		if err != nil {
			// v3 often identifies by tx_ref; try that next.
			if txRef := anyToString(event.Data["tx_ref"]); txRef != "" {
				ch, err = s.fwCli.GetCharge(ctx, s.defaultCreds, txRef)
			}
		}
		if err != nil {
			logger.WithError(err).Warn("verify charge failed — using webhook body")
		} else {
			return s.applyCharge(ctx, logger, ch)
		}
	}
	// Fallback: map from webhook body (v3 uses tx_ref; v4 uses reference).
	statusStr, _ := event.Data["status"].(string)
	ref := firstNonEmpty(anyToString(event.Data["reference"]), anyToString(event.Data["tx_ref"]))
	meta := extractMeta(event.Data)
	return s.pushPromptStatus(ctx, logger, chargeID, ref, statusStr, meta, event.Data)
}

func (s *FlutterwaveWebhookServer) canVerify() bool {
	if s.defaultCreds == nil {
		return false
	}
	if client.IsV3Credentials(s.defaultCreds) {
		return s.defaultCreds.SecretKey != "" || s.defaultCreds.ClientSecret != ""
	}
	return s.defaultCreds.ClientID != "" && s.defaultCreds.ClientSecret != ""
}

func anyToString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		// JSON numbers decode as float64
		if t == float64(int64(t)) {
			return strings.TrimSpace(strings.ReplaceAll(
				// avoid scientific notation for IDs
				fmt.Sprintf("%.0f", t), " ", ""))
		}
		return fmt.Sprintf("%v", t)
	case json.Number:
		return t.String()
	case int, int64, int32:
		return fmt.Sprintf("%v", t)
	default:
		return ""
	}
}

func (s *FlutterwaveWebhookServer) applyCharge(
	ctx context.Context,
	logger *util.LogEntry,
	ch *client.Charge,
) error {
	if ch == nil {
		return nil
	}
	meta := map[string]string{}
	for k, v := range ch.Meta {
		if s, ok := v.(string); ok {
			meta[k] = s
		}
	}
	return s.pushPromptStatus(ctx, logger, ch.ID, ch.Reference, ch.Status, meta, map[string]any{
		"amount":   ch.Amount,
		"currency": ch.Currency,
	})
}

func (s *FlutterwaveWebhookServer) pushPromptStatus(
	ctx context.Context,
	logger *util.LogEntry,
	chargeID, reference, statusStr string,
	meta map[string]string,
	extraFields map[string]any,
) error {
	promptID := meta["prompt_id"]
	if promptID == "" && strings.HasPrefix(reference, "prompt-") {
		promptID = strings.TrimPrefix(reference, "prompt-")
	}
	if promptID == "" {
		logger.WithField("reference", reference).Warn("charge webhook missing prompt_id")
		return nil
	}

	status := commonv1.STATUS_IN_PROCESS
	state := commonv1.STATE_ACTIVE
	switch strings.ToLower(statusStr) {
	case "succeeded", "successful", "success", "completed":
		status = commonv1.STATUS_SUCCESSFUL
	case "failed", "voided", "cancelled", "canceled":
		status = commonv1.STATUS_FAILED
		state = commonv1.STATE_INACTIVE
	}

	apiVersion := "v4"
	if s.defaultCreds != nil && client.IsV3Credentials(s.defaultCreds) {
		apiVersion = "v3"
	}
	extras := data.JSONMap{
		"entity_type":   "prompt",
		"provider":      "flutterwave",
		"api_version":   apiVersion,
		"charge_id":     chargeID,
		"reference":     reference,
		"charge_status": statusStr,
		"flw_event":     "charge.completed",
	}
	if amt, ok := extraFields["amount"].(float64); ok {
		extras["amount"] = amt
	}
	if cur, ok := extraFields["currency"].(string); ok {
		extras["currency"] = cur
	}

	externalID := chargeID
	if externalID == "" {
		externalID = reference
	}

	req := &commonv1.StatusUpdateRequest{
		Id:         promptID,
		State:      state,
		Status:     status,
		ExternalId: externalID,
		Extras:     extras.ToProtoStruct(),
	}
	if _, err := s.paymentCli.StatusUpdate(ctx, connect.NewRequest(req)); err != nil {
		logger.WithError(err).Error("StatusUpdate failed for charge")
		return err
	}
	return nil
}

func (s *FlutterwaveWebhookServer) handleTransferDisburse(
	ctx context.Context,
	logger *util.LogEntry,
	event *client.WebhookEvent,
) error {
	reference, _ := event.Data["reference"].(string)
	statusStr, _ := event.Data["status"].(string)
	transferID, _ := event.Data["id"].(string)

	paymentID := reference
	if strings.HasPrefix(reference, "pay-") {
		paymentID = strings.TrimPrefix(reference, "pay-")
	}
	// Also check meta.payment_id
	if meta := extractMeta(event.Data); meta["payment_id"] != "" {
		paymentID = meta["payment_id"]
	}
	if paymentID == "" {
		logger.Warn("transfer webhook missing reference")
		return nil
	}

	status := commonv1.STATUS_IN_PROCESS
	state := commonv1.STATE_ACTIVE
	switch strings.ToUpper(statusStr) {
	case "SUCCESSFUL", "SUCCESS", "SUCCEEDED":
		status = commonv1.STATUS_SUCCESSFUL
	case "FAILED", "CANCELLED", "CANCELED":
		status = commonv1.STATUS_FAILED
		state = commonv1.STATE_INACTIVE
	}

	extras := data.JSONMap{
		"entity_type":     "payment",
		"provider":        "flutterwave",
		"api_version":     "v4",
		"reference":       reference,
		"transfer_status": statusStr,
		"flw_event":       event.Type,
		"transfer_id":     transferID,
	}

	externalID := transferID
	if externalID == "" {
		externalID = reference
	}

	req := &commonv1.StatusUpdateRequest{
		Id:         paymentID,
		State:      state,
		Status:     status,
		ExternalId: externalID,
		Extras:     extras.ToProtoStruct(),
	}
	if _, err := s.paymentCli.StatusUpdate(ctx, connect.NewRequest(req)); err != nil {
		logger.WithError(err).Error("StatusUpdate failed for transfer")
		return err
	}
	return nil
}

func extractMeta(payload map[string]any) map[string]string {
	out := map[string]string{}
	raw, ok := payload["meta"]
	if !ok || raw == nil {
		return out
	}
	switch m := raw.(type) {
	case map[string]any:
		for k, v := range m {
			if s, ok := v.(string); ok {
				out[k] = s
			}
		}
	case map[string]string:
		return m
	}
	return out
}

func injectTenantFromQuery(ctx context.Context, r *http.Request) context.Context {
	return injectTenantContext(ctx, r.URL.Query().Get("tenant_id"), r.URL.Query().Get("partition_id"))
}

func injectTenantFromMeta(ctx context.Context, payload map[string]any) context.Context {
	meta := extractMeta(payload)
	return injectTenantContext(ctx, meta["tenant_id"], meta["partition_id"])
}

func injectTenantContext(ctx context.Context, tenantID, partitionID string) context.Context {
	if tenantID == "" && partitionID == "" {
		return ctx
	}
	claims := &security.AuthenticationClaims{
		TenantID:    tenantID,
		PartitionID: partitionID,
	}
	return claims.ClaimsToContext(ctx)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
