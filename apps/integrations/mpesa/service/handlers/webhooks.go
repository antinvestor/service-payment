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
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	"connectrpc.com/connect/v2"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/service/client"
	"github.com/antinvestor/service-payments/pkg/events"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/pitabwire/frame/v3/data"
	"github.com/pitabwire/frame/v3/security"
	"github.com/pitabwire/util"
)

// CredentialsResolver resolves Daraja credentials for a settings connection
// key (empty = service defaults). Used to re-query STK results.
type CredentialsResolver func(ctx context.Context, connection string) (*client.MpesaCredentials, error)

// maxWebhookBodyBytes bounds a callback body; Daraja callbacks are a few KB.
const maxWebhookBodyBytes = 1 << 20

// MpesaWebhookServer handles M-Pesa callback webhooks.
//
// Daraja does not sign callbacks, so a body is never trusted on its own:
// every handler looks the entity up in the payment service by the id our
// own callback URL (or the bill reference) carries, checks the stored
// binding and amount, and confirms with Daraja where a synchronous query
// exists (STK Push Query).
type MpesaWebhookServer struct {
	paymentCli   paymentv1connect.PaymentServiceClient
	mpesaCli     client.MpesaClient
	resolveCreds CredentialsResolver
	metrics      *integrationobs.Metrics
	now          func() time.Time
}

// NewMpesaWebhookServer creates a new webhook server. mpesaCli and
// resolveCreds are used to confirm STK success callbacks with Daraja's STK
// Push Query before a prompt is recorded SUCCESSFUL.
func NewMpesaWebhookServer(
	paymentCli paymentv1connect.PaymentServiceClient,
	mpesaCli client.MpesaClient,
	resolveCreds CredentialsResolver,
) *MpesaWebhookServer {
	return &MpesaWebhookServer{
		paymentCli:   paymentCli,
		mpesaCli:     mpesaCli,
		resolveCreds: resolveCreds,
		metrics:      integrationobs.NewMetrics("mpesa"),
		now:          time.Now,
	}
}

// NewRouterV1 creates the HTTP routes for M-Pesa webhooks.
func (s *MpesaWebhookServer) NewRouterV1() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/mpesa/stk", s.HandleSTKCallback)
	mux.HandleFunc("/webhook/mpesa/c2b/validation", s.HandleC2BValidation)
	mux.HandleFunc("/webhook/mpesa/c2b/confirmation", s.HandleC2BConfirmation)
	mux.HandleFunc("/webhook/mpesa/b2c", s.HandleB2CResult)
	mux.HandleFunc("/webhook/mpesa/b2c/timeout", s.HandleB2CTimeout)
	return mux
}

// errNotBound marks callbacks that do not match a request we issued; they
// must not change any status.
var errNotBound = errors.New("callback does not match an issued request")

// decode reads a bounded callback body.
func decode(w http.ResponseWriter, r *http.Request, v any) error {
	return json.NewDecoder(http.MaxBytesReader(w, r.Body, maxWebhookBodyBytes)).Decode(v)
}

// reject answers without recording anything: 409 when the callback does not
// match an issued request, otherwise 503 so Daraja retries.
func (s *MpesaWebhookServer) reject(ctx context.Context, w http.ResponseWriter, logger *util.LogEntry, kind string, err error) {
	if errors.Is(err, errNotBound) {
		logger.WithError(err).Warn("M-Pesa callback rejected")
		s.metrics.WebhookRejected(ctx, kind, "unknown_payment")
		http.Error(w, "callback does not match request", http.StatusConflict)
		return
	}
	logger.WithError(err).Error("could not verify M-Pesa callback")
	s.metrics.WebhookRejected(ctx, kind, "verification_failed")
	http.Error(w, "could not verify callback", http.StatusServiceUnavailable)
}

// record writes the status and answers Daraja.
func (s *MpesaWebhookServer) record(
	ctx context.Context,
	w http.ResponseWriter,
	logger *util.LogEntry,
	kind string,
	req *commonv1.StatusUpdateRequest,
) {
	req.State = events.StateFor(req.GetStatus())
	if _, err := s.paymentCli.StatusUpdate(ctx, req); err != nil {
		logger.WithError(err).Error("could not update payment status")
		s.metrics.WebhookRejected(ctx, kind, "status_update_error")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// loadStatus reads an entity's latest status extras. A missing record is
// reported as errNotBound; any other failure is retryable.
func (s *MpesaWebhookServer) loadStatus(ctx context.Context, entityID, entityType string) (data.JSONMap, error) {
	reqExtras := data.JSONMap{events.ExtraEntityType: entityType}
	resp, err := s.paymentCli.Status(ctx, &commonv1.StatusRequest{Id: entityID, Extras: reqExtras.ToProtoStruct()})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			return nil, fmt.Errorf("%w: no %s %q", errNotBound, entityType, entityID)
		}
		return nil, fmt.Errorf("load %s status: %w", entityType, err)
	}
	var stored data.JSONMap
	return stored.FromProtoStruct(resp.GetExtras()), nil
}

// --- STK push ---

// stkBinding is what the payment service already knows about the prompt.
type stkBinding struct {
	requestedAmount string
	connection      string
}

// HandleSTKCallback processes STK Push callback from M-Pesa.
//
// The prompt id comes from the callback URL (prompt_id) and must be bound to
// the same CheckoutRequestID in the prompt's stored status; a failure result
// is recorded only under that binding, and a success result only after
// Daraja's STK Push Query confirms it and the amount matches the one pushed.
func (s *MpesaWebhookServer) HandleSTKCallback(w http.ResponseWriter, r *http.Request) {
	ctx := injectTenantFromQuery(r.Context(), r)
	logger := util.Log(ctx).WithField("type", "mpesa.webhook.stk")
	defer logger.Release()

	s.metrics.WebhookReceived(ctx, "stk")

	var callback client.STKCallbackBody
	if err := decode(w, r, &callback); err != nil {
		logger.WithError(err).Error("failed to decode STK callback")
		s.metrics.WebhookRejected(ctx, "stk", "decode_error")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	stkCallback := callback.Body.StkCallback
	promptID := strings.TrimSpace(r.URL.Query().Get(client.CallbackParamPromptID))
	logger = logger.WithFields(map[string]any{
		"checkout_request_id": stkCallback.CheckoutRequestID,
		"prompt_id":           promptID,
	})

	if stkCallback.CheckoutRequestID == "" || promptID == "" {
		s.metrics.WebhookRejected(ctx, "stk", "missing_id")
		http.Error(w, "missing CheckoutRequestID or prompt_id", http.StatusBadRequest)
		return
	}

	binding, err := s.bindPrompt(ctx, promptID, stkCallback.CheckoutRequestID)
	if err != nil {
		s.reject(ctx, w, logger, "stk", err)
		return
	}

	extras := data.JSONMap{
		"result_code":                 strconv.Itoa(stkCallback.ResultCode),
		"result_desc":                 stkCallback.ResultDesc,
		"merchant_request_id":         stkCallback.MerchantRequestID,
		client.ExtraCheckoutRequestID: stkCallback.CheckoutRequestID,
		"prompt_id":                   promptID,
		events.ExtraEntityType:        events.EntityTypePrompt,
	}
	if binding.requestedAmount != "" {
		extras[client.ExtraRequestedAmount] = binding.requestedAmount
	}
	if binding.connection != "" {
		extras[client.ExtraCredentialsConnection] = binding.connection
	}
	callbackAmount := ""
	if stkCallback.CallbackMetadata != nil {
		for _, item := range stkCallback.CallbackMetadata.Item {
			extras[item.Label()] = fmt.Sprintf("%v", item.Value)
			if item.Label() == "Amount" {
				callbackAmount = fmt.Sprintf("%v", item.Value)
			}
		}
	}

	status := commonv1.STATUS_FAILED
	if stkCallback.ResultCode == 0 {
		status, err = s.confirmSTKSuccess(ctx, stkCallback.CheckoutRequestID, callbackAmount, binding, extras)
		if err != nil {
			logger.WithError(err).Error("could not confirm STK success with Daraja")
			s.metrics.WebhookRejected(ctx, "stk", "verification_failed")
			http.Error(w, "could not verify payment", http.StatusServiceUnavailable)
			return
		}
	} else {
		extras[events.ExtraError] = stkCallback.ResultDesc
	}

	s.record(ctx, w, logger, "stk", &commonv1.StatusUpdateRequest{
		Id:         promptID,
		Status:     status,
		ExternalId: stkCallback.MerchantRequestID,
		Extras:     extras.ToProtoStruct(),
	})
}

// bindPrompt checks that promptID's stored status was issued for
// checkoutRequestID. The STK worker records checkout_request_id on the
// prompt's IN_PROCESS status, and this handler carries it forward on every
// status it writes, so the latest status always holds the binding.
func (s *MpesaWebhookServer) bindPrompt(ctx context.Context, promptID, checkoutRequestID string) (*stkBinding, error) {
	stored, err := s.loadStatus(ctx, promptID, events.EntityTypePrompt)
	if err != nil {
		return nil, err
	}
	storedCheckoutID := stored.GetString(client.ExtraCheckoutRequestID)
	if storedCheckoutID == "" {
		return nil, fmt.Errorf("%w: prompt has no pending STK request", errNotBound)
	}
	if storedCheckoutID != checkoutRequestID {
		return nil, fmt.Errorf("%w: prompt was issued for a different CheckoutRequestID", errNotBound)
	}
	return &stkBinding{
		requestedAmount: stored.GetString(client.ExtraRequestedAmount),
		connection:      stored.GetString(client.ExtraCredentialsConnection),
	}, nil
}

// confirmSTKSuccess decides the status for a callback reporting ResultCode 0.
// It never returns SUCCESSFUL unless the amount matches the pushed amount
// (when known) and Daraja's STK Push Query also reports ResultCode 0. An error
// means the result could not be established and nothing should be recorded.
func (s *MpesaWebhookServer) confirmSTKSuccess(
	ctx context.Context,
	checkoutRequestID, callbackAmount string,
	binding *stkBinding,
	extras data.JSONMap,
) (commonv1.STATUS, error) {
	if binding.requestedAmount != "" && !amountsEqual(binding.requestedAmount, callbackAmount) {
		extras["verification"] = "amount_mismatch"
		extras[events.ExtraError] = "paid amount differs from the requested amount"
		return commonv1.STATUS_FAILED, nil
	}

	if s.mpesaCli == nil || s.resolveCreds == nil {
		return commonv1.STATUS_UNKNOWN, errors.New("STK query not configured")
	}
	creds, err := s.resolveCreds(ctx, binding.connection)
	if err != nil {
		return commonv1.STATUS_UNKNOWN, fmt.Errorf("resolve credentials: %w", err)
	}
	resp, err := s.mpesaCli.STKPushQuery(ctx, creds, client.NewSTKQueryRequest(creds, checkoutRequestID, s.now()))
	if err != nil {
		return commonv1.STATUS_UNKNOWN, fmt.Errorf("stk push query: %w", err)
	}

	resultCode := strings.TrimSpace(string(resp.ResultCode))
	extras["query_result_code"] = resultCode
	extras["query_result_desc"] = resp.ResultDesc
	switch resultCode {
	case "0":
		extras["verification"] = "stk_query"
		return commonv1.STATUS_SUCCESSFUL, nil
	case "":
		extras["verification"] = "stk_query_pending"
		return commonv1.STATUS_IN_PROCESS, nil
	default:
		extras["verification"] = "stk_query_mismatch"
		extras[events.ExtraError] = resp.ResultDesc
		return commonv1.STATUS_FAILED, nil
	}
}

// amountsEqual compares decimal amount strings (e.g. "100" and "100.00").
func amountsEqual(a, b string) bool {
	x, errA := strconv.ParseFloat(strings.TrimSpace(a), 64)
	y, errB := strconv.ParseFloat(strings.TrimSpace(b), 64)
	if errA != nil || errB != nil {
		return false
	}
	return math.Abs(x-y) < amountTolerance
}

// amountTolerance absorbs float formatting of whole-shilling amounts.
const amountTolerance = 0.005

// --- C2B ---

// HandleC2BValidation processes C2B validation requests. Every transaction
// is accepted (Daraja would otherwise reverse the customer's payment); the
// confirmation handler decides whether it settles anything.
func (s *MpesaWebhookServer) HandleC2BValidation(w http.ResponseWriter, r *http.Request) {
	ctx := injectTenantFromQuery(r.Context(), r)
	logger := util.Log(ctx).WithField("type", "mpesa.webhook.c2b.validation")
	defer logger.Release()

	s.metrics.WebhookReceived(ctx, "c2b_validation")

	var req client.C2BValidationRequest
	if err := decode(w, r, &req); err != nil {
		logger.WithError(err).Error("failed to decode C2B validation")
		s.metrics.WebhookRejected(ctx, "c2b_validation", "decode_error")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	logger.WithFields(map[string]any{
		"trans_id":        req.TransID,
		"bill_ref_number": req.BillRefNumber,
		"trans_amount":    req.TransAmount,
		"msisdn":          req.MSISDN,
	}).Info("C2B validation accepted")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"ResultCode": "0", "ResultDesc": "Accepted"})
}

// HandleC2BConfirmation processes C2B confirmation callbacks.
//
// The body is unauthenticated, so a confirmation only settles a prompt (or
// payment) whose status the payment service already holds under the
// BillRefNumber and whose requested amount and currency match what Daraja
// reports; anything else is answered 409 and recorded nowhere.
func (s *MpesaWebhookServer) HandleC2BConfirmation(w http.ResponseWriter, r *http.Request) {
	ctx := injectTenantFromQuery(r.Context(), r)
	logger := util.Log(ctx).WithField("type", "mpesa.webhook.c2b.confirmation")
	defer logger.Release()

	s.metrics.WebhookReceived(ctx, "c2b_confirmation")

	var req client.C2BValidationRequest
	if err := decode(w, r, &req); err != nil {
		logger.WithError(err).Error("failed to decode C2B confirmation")
		s.metrics.WebhookRejected(ctx, "c2b_confirmation", "decode_error")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	billRef := strings.TrimSpace(req.BillRefNumber)
	logger = logger.WithFields(map[string]any{"trans_id": req.TransID, "bill_ref_number": billRef})
	if billRef == "" {
		s.metrics.WebhookRejected(ctx, "c2b_confirmation", "missing_id")
		http.Error(w, "missing BillRefNumber", http.StatusBadRequest)
		return
	}

	entityType, stored, err := s.loadC2BTarget(ctx, billRef)
	if err != nil {
		s.reject(ctx, w, logger, "c2b_confirmation", err)
		return
	}
	if err = matchesStored(stored, req.TransAmount, client.Currency); err != nil {
		logger.WithFields(map[string]any{
			"requested_amount": stored.GetString(client.ExtraRequestedAmount),
			"trans_amount":     req.TransAmount,
		}).Warn("C2B confirmation amount does not match the request")
		s.reject(ctx, w, logger, "c2b_confirmation", err)
		return
	}

	extras := data.JSONMap{
		"trans_id":             req.TransID,
		"trans_time":           req.TransTime,
		"trans_amount":         req.TransAmount,
		"bill_ref_number":      billRef,
		"msisdn":               req.MSISDN,
		"first_name":           req.FirstName,
		"middle_name":          req.MiddleName,
		"last_name":            req.LastName,
		"org_account_balance":  req.OrgAccountBalance,
		"verification":         "c2b_binding_amount",
		events.ExtraEntityType: entityType,
		"amount":               req.TransAmount,
		"currency":             client.Currency,
	}
	carryBinding(extras, stored)

	s.record(ctx, w, logger, "c2b_confirmation", &commonv1.StatusUpdateRequest{
		Id:         billRef,
		Status:     commonv1.STATUS_SUCCESSFUL,
		ExternalId: req.TransID,
		Extras:     extras.ToProtoStruct(),
	})
}

// loadC2BTarget finds the entity a BillRefNumber refers to: a prompt first,
// then a payment.
func (s *MpesaWebhookServer) loadC2BTarget(ctx context.Context, billRef string) (string, data.JSONMap, error) {
	var lastErr error
	for _, entityType := range []string{events.EntityTypePrompt, events.EntityTypePayment} {
		stored, err := s.loadStatus(ctx, billRef, entityType)
		if err == nil {
			return entityType, stored, nil
		}
		if !errors.Is(err, errNotBound) {
			return "", nil, err
		}
		lastErr = err
	}
	return "", nil, lastErr
}

// matchesStored checks a callback amount/currency against the request
// recorded on the entity's status. A request issued without a recorded
// amount cannot be checked and is refused.
func matchesStored(stored data.JSONMap, amount, currency string) error {
	want := stored.GetString(client.ExtraRequestedAmount)
	if want == "" {
		return fmt.Errorf("%w: no requested amount recorded", errNotBound)
	}
	if !amountsEqual(want, amount) {
		return fmt.Errorf("%w: amount %q differs from requested %q", errNotBound, amount, want)
	}
	if wantCcy := stored.GetString(client.ExtraRequestedCurrency); wantCcy != "" && !strings.EqualFold(wantCcy, currency) {
		return fmt.Errorf("%w: currency %q differs from requested %q", errNotBound, currency, wantCcy)
	}
	return nil
}

// carryBinding keeps the request binding on every status written, so the
// latest status always holds it.
func carryBinding(extras, stored data.JSONMap) {
	for _, key := range []string{
		client.ExtraCheckoutRequestID, client.ExtraOriginatorConversationID,
		client.ExtraRequestedAmount, client.ExtraRequestedCurrency, client.ExtraCredentialsConnection,
	} {
		if v := stored.GetString(key); v != "" {
			extras[key] = v
		}
	}
}

// --- B2C ---

// HandleB2CResult processes B2C result callbacks.
//
// The payment id comes from our own result URL (payment_id); the callback
// is only honoured when the payment's stored status carries the same
// OriginatorConversationID the payout was issued with, and a success is
// recorded only when the paid TransactionAmount equals the amount requested.
//
// Daraja's Transaction Status API cannot confirm this synchronously: it only
// acknowledges the query and posts the status to a ResultURL later (see
// client.TransactionStatus). The recorded verification is therefore the
// binding plus the amount check ("b2c_binding_amount").
func (s *MpesaWebhookServer) HandleB2CResult(w http.ResponseWriter, r *http.Request) {
	ctx := injectTenantFromQuery(r.Context(), r)
	logger := util.Log(ctx).WithField("type", "mpesa.webhook.b2c")
	defer logger.Release()

	s.metrics.WebhookReceived(ctx, "b2c")

	result, paymentID, ok := s.decodeB2C(ctx, w, r, logger, "b2c")
	if !ok {
		return
	}
	stored, err := s.bindPayment(ctx, paymentID, result.OriginatorConversationID)
	if err != nil {
		s.reject(ctx, w, logger, "b2c", err)
		return
	}

	extras := data.JSONMap{
		"result_code":                        strconv.Itoa(result.ResultCode),
		"result_desc":                        result.ResultDesc,
		"conversation_id":                    result.ConversationID,
		client.ExtraOriginatorConversationID: result.OriginatorConversationID,
		"transaction_id":                     result.TransactionID,
		events.ExtraEntityType:               events.EntityTypePayment,
	}
	carryBinding(extras, stored)
	paidAmount := ""
	if result.ResultParameters != nil {
		for _, param := range result.ResultParameters.ResultParameter {
			extras[param.Label()] = fmt.Sprintf("%v", param.Value)
			if param.Label() == client.ResultParamTransactionAmount {
				paidAmount = fmt.Sprintf("%v", param.Value)
			}
		}
	}

	status := commonv1.STATUS_FAILED
	switch {
	case result.ResultCode != 0:
		extras["verification"] = "b2c_binding"
		extras[events.ExtraError] = result.ResultDesc
	case matchesStored(stored, paidAmount, client.Currency) != nil:
		logger.WithFields(map[string]any{
			"requested_amount": stored.GetString(client.ExtraRequestedAmount),
			"paid_amount":      paidAmount,
		}).Error("Daraja paid a different amount than requested")
		extras["verification"] = "amount_mismatch"
		extras[events.ExtraError] = "paid amount differs from the requested amount"
	default:
		status = commonv1.STATUS_SUCCESSFUL
		extras["verification"] = "b2c_binding_amount"
		extras["amount"] = paidAmount
		extras["currency"] = client.Currency
	}

	s.record(ctx, w, logger, "b2c", &commonv1.StatusUpdateRequest{
		Id:         paymentID,
		Status:     status,
		ExternalId: result.TransactionID,
		Extras:     extras.ToProtoStruct(),
	})
}

// HandleB2CTimeout processes B2C queue timeout callbacks. A timeout says
// nothing about the payout's outcome, so a bound payment is left IN_PROCESS
// until Daraja's result arrives (or a reconciliation settles it); nothing is
// recorded as failed on a timeout alone.
func (s *MpesaWebhookServer) HandleB2CTimeout(w http.ResponseWriter, r *http.Request) {
	ctx := injectTenantFromQuery(r.Context(), r)
	logger := util.Log(ctx).WithField("type", "mpesa.webhook.b2c.timeout")
	defer logger.Release()

	s.metrics.WebhookReceived(ctx, "b2c_timeout")

	result, paymentID, ok := s.decodeB2C(ctx, w, r, logger, "b2c_timeout")
	if !ok {
		return
	}
	if _, err := s.bindPayment(ctx, paymentID, result.OriginatorConversationID); err != nil {
		s.reject(ctx, w, logger, "b2c_timeout", err)
		return
	}
	logger.WithFields(map[string]any{
		"result_code": result.ResultCode,
		"result_desc": result.ResultDesc,
	}).Warn("B2C request timed out at Daraja; payment left in process pending the result")
	w.WriteHeader(http.StatusOK)
}

// decodeB2C reads a B2C callback and the payment id from its URL.
func (s *MpesaWebhookServer) decodeB2C(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	logger *util.LogEntry,
	kind string,
) (*client.B2CResult, string, bool) {
	var callback client.B2CCallbackBody
	if err := decode(w, r, &callback); err != nil {
		logger.WithError(err).Error("failed to decode B2C callback")
		s.metrics.WebhookRejected(ctx, kind, "decode_error")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return nil, "", false
	}
	paymentID := strings.TrimSpace(r.URL.Query().Get(client.CallbackParamPaymentID))
	logger.WithFields(map[string]any{
		"payment_id":                 paymentID,
		"conversation_id":            callback.Result.ConversationID,
		"originator_conversation_id": callback.Result.OriginatorConversationID,
	})
	if paymentID == "" || strings.TrimSpace(callback.Result.OriginatorConversationID) == "" {
		s.metrics.WebhookRejected(ctx, kind, "missing_id")
		http.Error(w, "missing payment_id or OriginatorConversationID", http.StatusBadRequest)
		return nil, "", false
	}
	return &callback.Result, paymentID, true
}

// bindPayment checks that paymentID's stored status was issued with
// originatorConversationID (the payout worker records it before sending).
func (s *MpesaWebhookServer) bindPayment(ctx context.Context, paymentID, originatorConversationID string) (data.JSONMap, error) {
	stored, err := s.loadStatus(ctx, paymentID, events.EntityTypePayment)
	if err != nil {
		return nil, err
	}
	issued := stored.GetString(client.ExtraOriginatorConversationID)
	if issued == "" {
		return nil, fmt.Errorf("%w: payment has no B2C request recorded", errNotBound)
	}
	if issued != strings.TrimSpace(originatorConversationID) {
		return nil, fmt.Errorf("%w: payment was issued with a different OriginatorConversationID", errNotBound)
	}
	return stored, nil
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
