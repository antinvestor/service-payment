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
	"net/http"
	"strconv"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/mpesa/service/client"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/frame/v2/security"
	"github.com/pitabwire/util"
)

// MpesaWebhookServer handles M-Pesa callback webhooks.
type MpesaWebhookServer struct {
	paymentCli paymentv1connect.PaymentServiceClient
	metrics    *integrationobs.Metrics
}

// NewMpesaWebhookServer creates a new webhook server.
func NewMpesaWebhookServer(paymentCli paymentv1connect.PaymentServiceClient) *MpesaWebhookServer {
	return &MpesaWebhookServer{
		paymentCli: paymentCli,
		metrics:    integrationobs.NewMetrics("mpesa"),
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

// HandleSTKCallback processes STK Push callback from M-Pesa.
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
	logger = logger.WithField("checkout_request_id", stkCallback.CheckoutRequestID)

	status := commonv1.STATUS_SUCCESSFUL
	state := commonv1.STATE_ACTIVE
	if stkCallback.ResultCode != 0 {
		status = commonv1.STATUS_FAILED
		state = commonv1.STATE_INACTIVE
	}

	extras := data.JSONMap{
		"result_code":         strconv.Itoa(stkCallback.ResultCode),
		"result_desc":         stkCallback.ResultDesc,
		"merchant_request_id": stkCallback.MerchantRequestID,
		"checkout_request_id": stkCallback.CheckoutRequestID,
		"entity_type":         "prompt",
	}

	if stkCallback.CallbackMetadata != nil {
		for _, item := range stkCallback.CallbackMetadata.Item {
			extras[item.Name] = fmt.Sprintf("%v", item.Value)
		}
	}

	statusReq := &commonv1.StatusUpdateRequest{
		Id:         stkCallback.CheckoutRequestID,
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
