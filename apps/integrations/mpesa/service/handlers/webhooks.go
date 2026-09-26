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
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/service/client"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/frame/v2/security"
	"github.com/pitabwire/util"
)

// CredentialsResolver resolves Daraja credentials for a settings connection
// key (empty = service defaults). Used to re-query STK results.
type CredentialsResolver func(ctx context.Context, connection string) (*client.MpesaCredentials, error)

// MpesaWebhookServer handles M-Pesa callback webhooks.
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

// errSTKRejected marks callbacks that must not change any status.
var errSTKRejected = errors.New("stk callback rejected")

// stkBinding is what the payment service already knows about the prompt.
type stkBinding struct {
	requestedAmount string
	connection      string
}

// HandleSTKCallback processes STK Push callback from M-Pesa.
//
// The callback body is unauthenticated (Daraja does not sign callbacks), so:
//   - the prompt id comes from the callback URL (prompt_id) and must be bound
//     to the same CheckoutRequestID in the prompt's stored status;
//   - a success result is only recorded after Daraja's STK Push Query confirms
//     it and the callback amount matches the amount pushed.
//
// Callbacks without prompt_id (STK pushes issued before this parameter
// existed) keep the legacy behaviour of recording under CheckoutRequestID,
// but success is still confirmed with Daraja.
func (s *MpesaWebhookServer) HandleSTKCallback(w http.ResponseWriter, r *http.Request) {
	ctx := injectTenantFromQuery(r.Context(), r)
	logger := util.Log(ctx).WithField("type", "mpesa.webhook.stk")
	defer logger.Release()

	s.metrics.WebhookReceived(ctx, "stk")

	var callback client.STKCallbackBody
	if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
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

	if stkCallback.CheckoutRequestID == "" {
		s.metrics.WebhookRejected(ctx, "stk", "missing_checkout_request_id")
		http.Error(w, "missing CheckoutRequestID", http.StatusBadRequest)
		return
	}

	extras := data.JSONMap{
		"result_code":                 strconv.Itoa(stkCallback.ResultCode),
		"result_desc":                 stkCallback.ResultDesc,
		"merchant_request_id":         stkCallback.MerchantRequestID,
		client.ExtraCheckoutRequestID: stkCallback.CheckoutRequestID,
		"entity_type":                 "prompt",
	}
	callbackAmount := ""
	if stkCallback.CallbackMetadata != nil {
		for _, item := range stkCallback.CallbackMetadata.Item {
			extras[item.Name] = fmt.Sprintf("%v", item.Value)
			if item.Name == "Amount" {
				callbackAmount = fmt.Sprintf("%v", item.Value)
			}
		}
	}

	entityID := stkCallback.CheckoutRequestID
	binding := &stkBinding{}
	if promptID != "" {
		var err error
		binding, err = s.bindPrompt(ctx, promptID, stkCallback.CheckoutRequestID)
		if err != nil {
			logger.WithError(err).Warn("STK callback does not match a pending prompt")
			s.metrics.WebhookRejected(ctx, "stk", "prompt_binding")
			code := http.StatusServiceUnavailable
			if errors.Is(err, errSTKRejected) {
				code = http.StatusConflict
			}
			http.Error(w, "callback does not match prompt", code)
			return
		}
		entityID = promptID
		extras["prompt_id"] = promptID
		if binding.requestedAmount != "" {
			extras[client.ExtraRequestedAmount] = binding.requestedAmount
		}
		if binding.connection != "" {
			extras[client.ExtraCredentialsConnection] = binding.connection
		}
	}

	status := commonv1.STATUS_FAILED
	state := commonv1.STATE_INACTIVE
	if stkCallback.ResultCode == 0 {
		var err error
		status, err = s.confirmSTKSuccess(ctx, stkCallback.CheckoutRequestID, callbackAmount, binding, extras)
		if err != nil {
			logger.WithError(err).Error("could not confirm STK success with Daraja")
			s.metrics.WebhookRejected(ctx, "stk", "verification_error")
			http.Error(w, "could not verify payment", http.StatusServiceUnavailable)
			return
		}
		state = commonv1.STATE_ACTIVE
		if status == commonv1.STATUS_FAILED {
			state = commonv1.STATE_INACTIVE
		}
	}

	statusReq := &commonv1.StatusUpdateRequest{
		Id:         entityID,
		State:      state,
		Status:     status,
		ExternalId: stkCallback.MerchantRequestID,
		Extras:     extras.ToProtoStruct(),
	}

	if _, err := s.paymentCli.StatusUpdate(ctx, connect.NewRequest(statusReq)); err != nil {
		logger.WithError(err).Error("could not update payment status")
		s.metrics.WebhookRejected(ctx, "stk", "status_update_error")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// bindPrompt checks that promptID's stored status was issued for
// checkoutRequestID. The STK worker records checkout_request_id on the
// prompt's IN_PROCESS status, and this handler carries it forward on every
// status it writes, so the latest status always holds the binding.
func (s *MpesaWebhookServer) bindPrompt(
	ctx context.Context,
	promptID, checkoutRequestID string,
) (*stkBinding, error) {
	statusExtras := data.JSONMap{"entity_type": "prompt"}
	resp, err := s.paymentCli.Status(ctx, connect.NewRequest(&commonv1.StatusRequest{
		Id:     promptID,
		Extras: statusExtras.ToProtoStruct(),
	}))
	if err != nil {
		return nil, fmt.Errorf("load prompt status: %w", err)
	}
	var stored data.JSONMap
	stored = stored.FromProtoStruct(resp.Msg.GetExtras())
	storedCheckoutID := stored.GetString(client.ExtraCheckoutRequestID)
	if storedCheckoutID == "" {
		return nil, fmt.Errorf("%w: prompt has no pending STK request", errSTKRejected)
	}
	if storedCheckoutID != checkoutRequestID {
		return nil, fmt.Errorf("%w: prompt was issued for a different CheckoutRequestID", errSTKRejected)
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

// HandleC2BValidation processes C2B validation requests.
func (s *MpesaWebhookServer) HandleC2BValidation(w http.ResponseWriter, r *http.Request) {
	ctx := injectTenantFromQuery(r.Context(), r)
	logger := util.Log(ctx).WithField("type", "mpesa.webhook.c2b.validation")
	defer logger.Release()

	s.metrics.WebhookReceived(ctx, "c2b_validation")

	var req client.C2BValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("failed to decode C2B validation")
		s.metrics.WebhookRejected(ctx, "c2b_validation", "decode_error")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	logger.WithFields(map[string]any{"trans_id": req.TransID, "msisdn": req.MSISDN}).Debug("C2B validation received")

	// Accept the transaction
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"ResultCode": "0", "ResultDesc": "Accepted"})
}

// HandleC2BConfirmation processes C2B confirmation callbacks.
func (s *MpesaWebhookServer) HandleC2BConfirmation(w http.ResponseWriter, r *http.Request) {
	ctx := injectTenantFromQuery(r.Context(), r)
	logger := util.Log(ctx).WithField("type", "mpesa.webhook.c2b.confirmation")
	defer logger.Release()

	s.metrics.WebhookReceived(ctx, "c2b_confirmation")

	var req client.C2BValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("failed to decode C2B confirmation")
		s.metrics.WebhookRejected(ctx, "c2b_confirmation", "decode_error")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	logger.WithField("trans_id", req.TransID).Debug("C2B confirmation received")

	extras := data.JSONMap{
		"trans_id":            req.TransID,
		"trans_time":          req.TransTime,
		"trans_amount":        req.TransAmount,
		"bill_ref_number":     req.BillRefNumber,
		"msisdn":              req.MSISDN,
		"first_name":          req.FirstName,
		"middle_name":         req.MiddleName,
		"last_name":           req.LastName,
		"org_account_balance": req.OrgAccountBalance,
		"entity_type":         "payment",
	}

	statusReq := &commonv1.StatusUpdateRequest{
		Id:         req.BillRefNumber,
		State:      commonv1.STATE_ACTIVE,
		Status:     commonv1.STATUS_SUCCESSFUL,
		ExternalId: req.TransID,
		Extras:     extras.ToProtoStruct(),
	}

	if _, err := s.paymentCli.StatusUpdate(ctx, connect.NewRequest(statusReq)); err != nil {
		logger.WithError(err).Error("could not update payment status for C2B confirmation")
		s.metrics.WebhookRejected(ctx, "c2b_confirmation", "status_update_error")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleB2CResult processes B2C result callbacks.
func (s *MpesaWebhookServer) HandleB2CResult(w http.ResponseWriter, r *http.Request) {
	ctx := injectTenantFromQuery(r.Context(), r)
	logger := util.Log(ctx).WithField("type", "mpesa.webhook.b2c")
	defer logger.Release()

	s.metrics.WebhookReceived(ctx, "b2c")

	var callback client.B2CCallbackBody
	if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
		logger.WithError(err).Error("failed to decode B2C callback")
		s.metrics.WebhookRejected(ctx, "b2c", "decode_error")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result := callback.Result
	logger = logger.WithField("conversation_id", result.ConversationID)

	status := commonv1.STATUS_SUCCESSFUL
	state := commonv1.STATE_ACTIVE
	if result.ResultCode != 0 {
		status = commonv1.STATUS_FAILED
		state = commonv1.STATE_INACTIVE
	}

	extras := data.JSONMap{
		"result_code":                strconv.Itoa(result.ResultCode),
		"result_desc":                result.ResultDesc,
		"conversation_id":            result.ConversationID,
		"originator_conversation_id": result.OriginatorConversationID,
		"transaction_id":             result.TransactionID,
		"entity_type":                "payment",
	}

	if result.ResultParameters != nil {
		for _, param := range result.ResultParameters.ResultParameter {
			extras[param.Name] = fmt.Sprintf("%v", param.Value)
		}
	}

	statusReq := &commonv1.StatusUpdateRequest{
		Id:         result.OriginatorConversationID,
		State:      state,
		Status:     status,
		ExternalId: result.TransactionID,
		Extras:     extras.ToProtoStruct(),
	}

	if _, err := s.paymentCli.StatusUpdate(ctx, connect.NewRequest(statusReq)); err != nil {
		logger.WithError(err).Error("could not update payment status for B2C result")
		s.metrics.WebhookRejected(ctx, "b2c", "status_update_error")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleB2CTimeout processes B2C timeout callbacks.
func (s *MpesaWebhookServer) HandleB2CTimeout(w http.ResponseWriter, r *http.Request) {
	ctx := injectTenantFromQuery(r.Context(), r)
	logger := util.Log(ctx).WithField("type", "mpesa.webhook.b2c.timeout")
	defer logger.Release()

	s.metrics.WebhookReceived(ctx, "b2c_timeout")

	var callback client.B2CCallbackBody
	if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
		logger.WithError(err).Error("failed to decode B2C timeout callback")
		s.metrics.WebhookRejected(ctx, "b2c_timeout", "decode_error")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result := callback.Result

	extras := data.JSONMap{
		"result_code":                strconv.Itoa(result.ResultCode),
		"result_desc":                result.ResultDesc,
		"originator_conversation_id": result.OriginatorConversationID,
		"timeout":                    "true",
		"entity_type":                "payment",
	}

	statusReq := &commonv1.StatusUpdateRequest{
		Id:         result.OriginatorConversationID,
		State:      commonv1.STATE_ACTIVE,
		Status:     commonv1.STATUS_UNKNOWN,
		ExternalId: result.ConversationID,
		Extras:     extras.ToProtoStruct(),
	}

	if _, err := s.paymentCli.StatusUpdate(ctx, connect.NewRequest(statusReq)); err != nil {
		logger.WithError(err).Error("could not update payment status for B2C timeout")
		s.metrics.WebhookRejected(ctx, "b2c_timeout", "status_update_error")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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
