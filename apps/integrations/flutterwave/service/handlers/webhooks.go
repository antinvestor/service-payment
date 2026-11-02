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
	"crypto/hmac"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/config"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/queue"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/pitabwire/frame/v3/security"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/types/known/structpb"
)

// maxWebhookBody bounds a webhook body; Flutterwave events are a few KB.
const maxWebhookBody = 1 << 20

// errUnverifiable is an event whose record Flutterwave could not be asked
// for right now: answered with a retryable status so Flutterwave resends.
var errUnverifiable = errors.New("flutterwave record could not be verified")

// FlutterwaveWebhookServer handles Flutterwave webhooks.
// Docs: https://developer.flutterwave.com/docs/webhooks
//
// Nothing in a webhook body is trusted: the signature proves only that
// Flutterwave sent it, so the charge or transfer it names is re-read from
// the API and only that record — its status, amount and meta (prompt,
// payment and tenancy) — is reported.
type FlutterwaveWebhookServer struct {
	paymentCli    paymentv1connect.PaymentServiceClient
	fwCli         client.FlutterwaveClient
	webhookSecret string
	creds         *client.Credentials
	metrics       *integrationobs.Metrics
}

// NewFlutterwaveWebhookServer constructs the HTTP webhook server for the
// platform's Flutterwave account.
func NewFlutterwaveWebhookServer(
	paymentCli paymentv1connect.PaymentServiceClient,
	fwCli client.FlutterwaveClient,
	cfg *config.FlutterwaveConfig,
) *FlutterwaveWebhookServer {
	creds := &client.Credentials{
		ClientID:      cfg.ClientID,
		ClientSecret:  cfg.ClientSecret,
		PublicKey:     firstNonEmpty(cfg.PublicKey, cfg.ClientID),
		SecretKey:     cfg.SecretKey,
		EncryptionKey: cfg.EncryptionKey,
		WebhookSecret: cfg.WebhookSecret,
		Environment:   cfg.Environment,
		OAuthTokenURL: cfg.OAuthTokenURL,
	}
	if strings.HasPrefix(creds.ClientSecret, "FLWSECK_") || strings.HasPrefix(creds.ClientSecret, "FLWSECK-") {
		creds.SecretKey = firstNonEmpty(creds.SecretKey, creds.ClientSecret)
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
		creds:         creds,
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

// verified reports whether the request carries a valid signature: v4
// flutterwave-signature = base64(HMAC-SHA256(secret, body)), or the v3
// verif-hash header equal to the secret (compared in constant time).
func (s *FlutterwaveWebhookServer) verified(r *http.Request, body []byte) bool {
	if sig := r.Header.Get("Flutterwave-Signature"); sig != "" {
		return s.fwCli.VerifyWebhookSignature(body, sig, s.webhookSecret)
	}
	if hash := r.Header.Get("Verif-Hash"); hash != "" && client.IsV3Credentials(s.creds) {
		return hmac.Equal([]byte(hash), []byte(s.webhookSecret))
	}
	return false
}

// HandleWebhook processes charge and transfer events.
func (s *FlutterwaveWebhookServer) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.metrics.WebhookReceived(ctx, "flutterwave")
	logger := util.Log(ctx).WithField("type", "flutterwave.webhook")
	defer logger.Release()

	if s.webhookSecret == "" {
		// Without a secret no event can be authenticated: refuse them all
		// (Flutterwave retries) rather than act on forgeable bodies.
		logger.Error("FLUTTERWAVE_WEBHOOK_SECRET is not set; webhook refused")
		s.metrics.WebhookRejected(ctx, "flutterwave", "not_configured")
		http.Error(w, "webhook not configured", http.StatusServiceUnavailable)
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxWebhookBody))
	if err != nil {
		s.metrics.WebhookRejected(ctx, "flutterwave", "decode_error")
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}
	if !s.verified(r, body) {
		logger.Warn("webhook signature verification failed")
		s.metrics.WebhookRejected(ctx, "flutterwave", "verification_failed")
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var event client.WebhookEvent
	if err = json.Unmarshal(body, &event); err != nil {
		s.metrics.WebhookRejected(ctx, "flutterwave", "decode_error")
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	eventType := strings.ToLower(event.EventType())
	logger = logger.WithField("event_type", eventType).WithField("webhook_id", event.WebhookID)

	switch eventType {
	case "charge.completed":
		err = s.handleCharge(ctx, logger, &event)
	case "transfer.disburse", "transfer.completed":
		err = s.handleTransfer(ctx, logger, &event)
	case "transfer.reversal":
		// Money that reached the payee came back. A payout already reported
		// successful is never re-marked; the reversal needs a ledger entry
		// and a person to look at it.
		logger.WithField("transfer_id", anyToString(event.Data["id"])).
			WithField("reference", anyToString(event.Data["reference"])).
			Error("transfer reversed by Flutterwave; needs manual reconciliation")
		s.metrics.WebhookRejected(ctx, "flutterwave", "transfer_reversal")
	default:
		logger.Debug("unhandled event type")
	}

	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	case errors.Is(err, errUnverifiable):
		s.metrics.WebhookRejected(ctx, "flutterwave", "verify_unavailable")
		logger.WithError(err).Warn("webhook not yet verifiable; Flutterwave will retry")
		http.Error(w, "verification unavailable", http.StatusBadGateway)
	default:
		s.metrics.WebhookRejected(ctx, "flutterwave", "status_update_error")
		logger.WithError(err).Error("webhook processing error")
		http.Error(w, "processing error", http.StatusInternalServerError)
	}
}

// HandleReturn is the browser landing after a redirect next_action. It
// only tells the payer to go back: the outcome arrives by webhook and is
// confirmed by checkout, never inferred from a browser visit.
func (s *FlutterwaveWebhookServer) HandleReturn(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<!doctype html><html><body style="font-family:system-ui;padding:2rem">
<h1>Thank you</h1>
<p>You can close this window. The merchant will confirm your payment shortly.</p>
</body></html>`))
}

func (s *FlutterwaveWebhookServer) handleCharge(ctx context.Context, logger *util.LogEntry, event *client.WebhookEvent) error {
	chargeID := anyToString(event.Data["id"])
	if chargeID == "" {
		logger.Warn("charge event without id ignored")
		return nil
	}
	ch, err := s.fwCli.GetCharge(ctx, s.creds, chargeID)
	if err != nil {
		if client.IsNotFound(err) {
			logger.WithField("charge_id", chargeID).Warn("charge not found on this account; ignored")
			return nil
		}
		return fmt.Errorf("%w: get charge %s: %w", errUnverifiable, chargeID, err)
	}
	promptID := metaString(ch.Meta, queue.MetaPromptID)
	if promptID == "" {
		logger.WithField("charge_id", ch.ID).WithField("reference", ch.Reference).
			Warn("charge has no prompt; not ours")
		return nil
	}
	status, extras := queue.ChargeReport(ch, "webhook", queue.ExpectationFromMeta(ch.Meta))
	extras["flw_event"] = "charge.completed"
	return s.update(withTenancy(ctx, ch.Meta), promptID, ch.ID, status, extras)
}

func (s *FlutterwaveWebhookServer) handleTransfer(ctx context.Context, logger *util.LogEntry, event *client.WebhookEvent) error {
	transferID := anyToString(event.Data["id"])
	if transferID == "" {
		logger.Warn("transfer event without id ignored")
		return nil
	}
	if client.IsV3Credentials(s.creds) {
		// v3 transfers cannot be re-read here; production runs on v4 OAuth.
		return fmt.Errorf("%w: v3 transfer %s cannot be re-read", errUnverifiable, transferID)
	}
	tr, err := s.fwCli.GetTransfer(ctx, s.creds, transferID)
	if err != nil {
		if client.IsNotFound(err) {
			logger.WithField("transfer_id", transferID).Warn("transfer not found on this account; ignored")
			return nil
		}
		return fmt.Errorf("%w: get transfer %s: %w", errUnverifiable, transferID, err)
	}
	paymentID := metaString(tr.Meta, "payment_id")
	if paymentID == "" {
		logger.WithField("transfer_id", tr.ID).WithField("reference", tr.Reference).
			Warn("transfer has no payment; not ours")
		return nil
	}
	extras := map[string]any{
		"entity_type":     "payment",
		"provider":        "flutterwave",
		"api_version":     "v4",
		"reference":       tr.Reference,
		"transfer_id":     tr.ID,
		"transfer_status": tr.Status,
		"flw_event":       event.EventType(),
	}
	if tr.DestinationCurrency != "" {
		extras["currency"] = tr.DestinationCurrency
	}
	if v, ok := tr.Amount["value"]; ok {
		extras["amount"] = anyToString(v)
	}
	return s.update(withTenancy(ctx, tr.Meta), paymentID, tr.ID, queue.TransferStatus(tr.Status), extras)
}

func (s *FlutterwaveWebhookServer) update(
	ctx context.Context,
	id, externalID string,
	status commonv1.STATUS,
	extras map[string]any,
) error {
	extra, err := structpb.NewStruct(extras)
	if err != nil {
		return err
	}
	state := commonv1.STATE_ACTIVE
	if status == commonv1.STATUS_FAILED {
		state = commonv1.STATE_INACTIVE
	}
	_, err = s.paymentCli.StatusUpdate(ctx, &commonv1.StatusUpdateRequest{
		Id:         id,
		State:      state,
		Status:     status,
		ExternalId: externalID,
		Extras:     extra,
	})
	return err
}

// withTenancy attributes the update to the tenancy recorded on the
// verified Flutterwave record — never to anything in the request.
func withTenancy(ctx context.Context, meta map[string]any) context.Context {
	tenantID := metaString(meta, queue.MetaTenantID)
	partitionID := metaString(meta, queue.MetaPartitionID)
	if tenantID == "" && partitionID == "" {
		return ctx
	}
	claims := &security.AuthenticationClaims{TenantID: tenantID, PartitionID: partitionID}
	return claims.ClaimsToContext(ctx)
}

func metaString(meta map[string]any, key string) string {
	return anyToString(meta[key])
}

func anyToString(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case json.Number:
		return t.String()
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	default:
		return ""
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
