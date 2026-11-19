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

// Package handlers receives Jenga callbacks. Jenga does not sign callbacks,
// so a callback is only ever allowed to settle a prompt this service raised:
// the prompt id and the per-prompt secret come from the callback URL the
// worker minted, the stored binding (secret, transaction reference and
// requested amount) must match, and nothing is ever recorded as an inbound
// payment from the body alone.
//
// The Jenga client offers no transaction query/status endpoint, so a success
// cannot be re-confirmed with Jenga before it is recorded. SUCCESSFUL is
// therefore recorded only when the secret matches (constant-time), the
// amount and currency match what was pushed, and the callback says success.
package handlers

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/events"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/pitabwire/frame/v3/data"
	"github.com/pitabwire/frame/v3/security"
	"github.com/pitabwire/util"
)

const (
	providerName = "jenga"
	// maxCallbackBody bounds an unauthenticated callback body (1 MiB).
	maxCallbackBody = 1 << 20

	kindSTK     = "stk"
	kindGeneral = "general"
)

// errUnbound marks callbacks that do not match a prompt this service raised.
var errUnbound = errors.New("callback is not bound to a pending prompt")

// errForged marks callbacks whose per-prompt secret does not match.
var errForged = errors.New("callback secret does not match the prompt")

// WebhookServer handles incoming callbacks from the Jenga API.
type WebhookServer struct {
	paymentCli paymentv1connect.PaymentServiceClient
	metrics    *integrationobs.Metrics
}

// NewWebhookServer creates a new webhook server for Jenga API callbacks.
func NewWebhookServer(paymentCli paymentv1connect.PaymentServiceClient) *WebhookServer {
	return &WebhookServer{
		paymentCli: paymentCli,
		metrics:    integrationobs.NewMetrics(providerName),
	}
}

// NewRouter creates the HTTP router for Jenga webhook endpoints.
func (ws *WebhookServer) NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST "+models.STKCallbackPath, ws.HandleStkCallback)
	mux.HandleFunc("POST "+models.GeneralCallbackPath, ws.HandleGeneralCallback)
	return mux
}

// binding is what the payment service recorded when the push was raised.
type binding struct {
	transactionRef    string
	requestedAmount   string
	requestedCurrency string
	connection        string
	transactionID     string
	secretHash        string
}

// settlement is what a callback reports about the bound prompt.
type settlement struct {
	transactionRef string
	amount         string
	currency       string
	success        bool
	failureCode    string
	failureMessage string
	extras         data.JSONMap
}

// HandleStkCallback processes STK push results from Jenga.
func (ws *WebhookServer) HandleStkCallback(w http.ResponseWriter, r *http.Request) {
	ctx := injectTenantFromQuery(r.Context(), r)
	ws.metrics.WebhookReceived(ctx, kindSTK)

	var callback models.StkCallback
	if !ws.decode(ctx, w, r, kindSTK, &callback) {
		return
	}
	if callback.Transaction == "" {
		ws.reject(ctx, w, kindSTK, "missing_id", "missing transactionReference", http.StatusBadRequest)
		return
	}
	raw, _ := json.Marshal(callback)
	s := &settlement{
		transactionRef: callback.Transaction,
		amount:         firstNonEmpty(callback.RequestAmount.String(), callback.DebitedAmount.String()),
		currency:       callback.Currency,
		success:        callback.Status,
		failureCode:    fmt.Sprintf("JENGA_%d", callback.Code),
		failureMessage: callback.Message,
		extras: data.JSONMap{
			"telco_reference": callback.Telco,
			"telco":           callback.TelcoName,
			"mobile_number":   callback.MobileNumber,
			"debited_amount":  callback.DebitedAmount.String(),
			"charge":          callback.Charge.String(),
			"message":         callback.Message,
			"response_code":   fmt.Sprintf("%d", callback.Code),
			"additional_info": string(raw),
		},
	}
	ws.settle(ctx, w, r, kindSTK, s)
}

// HandleGeneralCallback processes Jenga's generic payment notifications.
// Only notifications for a prompt this service raised are recorded; an
// unsolicited inbound payment cannot be bound and is refused.
func (ws *WebhookServer) HandleGeneralCallback(w http.ResponseWriter, r *http.Request) {
	ctx := injectTenantFromQuery(r.Context(), r)
	ws.metrics.WebhookReceived(ctx, kindGeneral)

	var callback models.CallbackRequest
	if !ws.decode(ctx, w, r, kindGeneral, &callback) {
		return
	}
	if callback.Transaction.Reference == "" {
		ws.reject(ctx, w, kindGeneral, "missing_id", "missing transaction reference", http.StatusBadRequest)
		return
	}
	raw, _ := json.Marshal(callback)
	s := &settlement{
		transactionRef: callback.Transaction.Reference,
		amount:         callback.Transaction.Amount.String(),
		currency:       callback.Transaction.Currency,
		success:        isSuccessStatus(callback.Transaction.Status),
		failureCode:    strings.ToUpper(strings.TrimSpace(callback.Transaction.Status)),
		failureMessage: firstNonEmpty(callback.Transaction.Remarks, callback.Transaction.Status),
		extras: data.JSONMap{
			"callback_type":   callback.CallbackType,
			"payment_mode":    callback.Transaction.PaymentMode,
			"bill_number":     callback.Transaction.BillNumber,
			"bank_reference":  callback.Bank.Reference,
			"mobile_number":   callback.Customer.MobileNumber,
			"charge":          callback.Transaction.ServiceCharge.String(),
			"provider_status": callback.Transaction.Status,
			"additional_info": string(raw),
		},
	}
	ws.settle(ctx, w, r, kindGeneral, s)
}

// decode reads a bounded callback body.
func (ws *WebhookServer) decode(ctx context.Context, w http.ResponseWriter, r *http.Request, kind string, into any) bool {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxCallbackBody))
	if err != nil {
		ws.reject(ctx, w, kind, "decode_error", "invalid request body", http.StatusBadRequest)
		return false
	}
	if err = json.Unmarshal(body, into); err != nil {
		util.Log(ctx).WithError(err).WithField("kind", kind).Error("failed to decode Jenga callback")
		ws.reject(ctx, w, kind, "decode_error", "invalid request body", http.StatusBadRequest)
		return false
	}
	return true
}

// settle binds the callback to the prompt named in the URL and records its
// outcome under that prompt. It never records a payment from the body.
func (ws *WebhookServer) settle(ctx context.Context, w http.ResponseWriter, r *http.Request, kind string, s *settlement) {
	promptID := strings.TrimSpace(r.URL.Query().Get(models.CallbackParamPromptID))
	logger := util.Log(ctx).WithFields(map[string]any{
		"handler":                  "jenga.webhook." + kind,
		"prompt_id":                promptID,
		models.ExtraTransactionRef: s.transactionRef,
		"amount":                   s.amount,
		"currency":                 s.currency,
	})
	defer logger.Release()

	if promptID == "" {
		logger.Warn("callback carries no prompt id; it cannot be bound to a prompt we raised")
		ws.reject(ctx, w, kind, "missing_id", "callback is not bound to a prompt", http.StatusConflict)
		return
	}

	secret := strings.TrimSpace(r.URL.Query().Get(models.CallbackParamSecret))
	b, err := ws.bindPrompt(ctx, promptID, secret, s.transactionRef)
	if err != nil {
		logger.WithError(err).Warn("callback does not match a pending prompt")
		switch {
		case errors.Is(err, errForged):
			ws.reject(ctx, w, kind, "verification_failed", "callback secret does not match", http.StatusForbidden)
		case errors.Is(err, errUnbound):
			ws.reject(ctx, w, kind, "unknown_payment", "callback does not match prompt", http.StatusConflict)
		default:
			ws.reject(ctx, w, kind, "verification_failed", "could not load prompt", http.StatusServiceUnavailable)
		}
		return
	}

	extras := s.extras
	extras[events.ExtraEntityType] = events.EntityTypePrompt
	extras[collection.ExtraProvider] = providerName
	extras["prompt_id"] = promptID
	extras[models.ExtraTransactionRef] = b.transactionRef
	extras[models.ExtraRequestedAmount] = b.requestedAmount
	extras[models.ExtraRequestedCurrency] = b.requestedCurrency
	if b.connection != "" {
		extras[models.ExtraCredentialsConnection] = b.connection
	}
	if b.transactionID != "" {
		extras[models.ExtraTransactionID] = b.transactionID
	}
	// The binding is carried forward so a later callback still binds.
	extras[models.ExtraCallbackSecretHash] = b.secretHash

	status := commonv1.STATUS_FAILED
	switch {
	case !s.success:
		extras["failure_code"] = s.failureCode
		extras["failure_message"] = firstNonEmpty(s.failureMessage, "The payment was not completed")
	case !amountsEqual(b.requestedAmount, s.amount) ||
		(b.requestedCurrency != "" && s.currency != "" && !strings.EqualFold(b.requestedCurrency, s.currency)):
		logger.Warn("callback amount differs from the amount pushed")
		extras["verification"] = "amount_mismatch"
		extras["failure_code"] = "AMOUNT_MISMATCH"
		extras["failure_message"] = "The amount paid differs from the amount requested"
	default:
		// No Jenga query endpoint exists in the client to re-confirm this;
		// the secret and amount checks above are the whole verification.
		status = commonv1.STATUS_SUCCESSFUL
		extras["amount"] = s.amount
		extras["currency"] = firstNonEmpty(s.currency, b.requestedCurrency)
	}

	if _, err = ws.paymentCli.StatusUpdate(ctx, &commonv1.StatusUpdateRequest{
		Id:         promptID,
		State:      events.StateFor(status),
		Status:     status,
		ExternalId: s.transactionRef,
		Extras:     extras.ToProtoStruct(),
	}); err != nil {
		logger.WithError(err).Error("failed to update prompt status from callback")
		ws.reject(ctx, w, kind, "status_update_error", "failed to process callback", http.StatusInternalServerError)
		return
	}

	logger.WithField("status", status.String()).Info("Jenga callback recorded")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Callback received"})
}

// bindPrompt loads promptID's stored status, checks the per-prompt secret
// before anything else, then checks the push was raised for transactionRef.
func (ws *WebhookServer) bindPrompt(ctx context.Context, promptID, secret, transactionRef string) (*binding, error) {
	statusExtras := data.JSONMap{events.ExtraEntityType: events.EntityTypePrompt}
	resp, err := ws.paymentCli.Status(ctx, &commonv1.StatusRequest{
		Id:     promptID,
		Extras: statusExtras.ToProtoStruct(),
	})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			return nil, fmt.Errorf("%w: prompt %s is unknown", errUnbound, promptID)
		}
		return nil, fmt.Errorf("load prompt status: %w", err)
	}
	var stored data.JSONMap
	stored = stored.FromProtoStruct(resp.GetExtras())
	storedHash := stored.GetString(models.ExtraCallbackSecretHash)
	if storedHash == "" || secret == "" ||
		subtle.ConstantTimeCompare([]byte(storedHash), []byte(models.HashCallbackSecret(secret))) != 1 {
		return nil, errForged
	}
	b := &binding{
		secretHash:        storedHash,
		transactionRef:    stored.GetString(models.ExtraTransactionRef),
		requestedAmount:   stored.GetString(models.ExtraRequestedAmount),
		requestedCurrency: stored.GetString(models.ExtraRequestedCurrency),
		connection:        stored.GetString(models.ExtraCredentialsConnection),
		transactionID:     stored.GetString(models.ExtraTransactionID),
	}
	if b.transactionRef == "" {
		return nil, fmt.Errorf("%w: prompt has no pending push", errUnbound)
	}
	if b.transactionRef != transactionRef {
		return nil, fmt.Errorf("%w: prompt was pushed with a different transaction reference", errUnbound)
	}
	return b, nil
}

func (ws *WebhookServer) reject(ctx context.Context, w http.ResponseWriter, kind, reason, message string, code int) {
	ws.metrics.WebhookRejected(ctx, kind, reason)
	http.Error(w, message, code)
}

// amountsEqual compares two decimal amounts exactly; an unparseable or
// missing amount never matches.
func amountsEqual(a, b string) bool {
	x, okA := new(big.Rat).SetString(strings.TrimSpace(a))
	y, okB := new(big.Rat).SetString(strings.TrimSpace(b))
	if !okA || !okB {
		return false
	}
	return x.Cmp(y) == 0
}

// isSuccessStatus reads Jenga's textual transaction status.
func isSuccessStatus(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "SUCCESS", "SUCCESSFUL", "COMPLETED", "PAID":
		return true
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// injectTenantFromQuery extracts tenant_id and partition_id from URL query params and injects into context.
func injectTenantFromQuery(ctx context.Context, r *http.Request) context.Context {
	tenantID := r.URL.Query().Get("tenant_id")
	partitionID := r.URL.Query().Get("partition_id")
	if tenantID == "" && partitionID == "" {
		return ctx
	}
	claims := &security.AuthenticationClaims{
		TenantID:    tenantID,
		PartitionID: partitionID,
	}
	return claims.ClaimsToContext(ctx)
}
