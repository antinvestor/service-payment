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
	"net/http"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/mtn/service/client"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/security"
	"github.com/pitabwire/util"
)

// MtnWebhookServer handles MTN MoMo callback webhooks.
type MtnWebhookServer struct {
	paymentCli paymentv1connect.PaymentServiceClient
	metrics    *integrationobs.Metrics
}

// NewMtnWebhookServer creates a new webhook server.
func NewMtnWebhookServer(paymentCli paymentv1connect.PaymentServiceClient) *MtnWebhookServer {
	return &MtnWebhookServer{
		paymentCli: paymentCli,
		metrics:    integrationobs.NewMetrics("mtn"),
	}
}

// NewRouterV1 creates the HTTP routes for MTN MoMo webhooks.
func (s *MtnWebhookServer) NewRouterV1() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/mtn/collection", s.HandleCollectionCallback)
	mux.HandleFunc("/webhook/mtn/disbursement", s.HandleDisbursementCallback)
	return mux
}

// HandleCollectionCallback processes requestToPay callbacks from MTN MoMo.
func (s *MtnWebhookServer) HandleCollectionCallback(w http.ResponseWriter, r *http.Request) {
	s.metrics.WebhookReceived(r.Context(), "collection")
	s.handleCallback(w, r, "mtn.webhook.collection", "prompt")
}

// HandleDisbursementCallback processes disbursement transfer callbacks from MTN MoMo.
func (s *MtnWebhookServer) HandleDisbursementCallback(w http.ResponseWriter, r *http.Request) {
	s.metrics.WebhookReceived(r.Context(), "disbursement")
	s.handleCallback(w, r, "mtn.webhook.disbursement", "payment")
}

// handleCallback is the shared logic for processing MTN MoMo callback webhooks.
func (s *MtnWebhookServer) handleCallback(w http.ResponseWriter, r *http.Request, logType, entityType string) {
	ctx := injectTenantFromQuery(r.Context(), r)
	logger := util.Log(ctx).WithField("type", logType)
	defer logger.Release()

	var callback client.CallbackBody
	if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
		logger.WithError(err).Error("failed to decode callback")
		s.metrics.WebhookRejected(ctx, entityType, "decode_error")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	logger = logger.WithField("external_id", callback.ExternalID)

	status := mapMtnStatus(callback.Status)
	state := commonv1.STATE_ACTIVE
	if status == commonv1.STATUS_FAILED {
		state = commonv1.STATE_INACTIVE
	}

	extras := data.JSONMap{
		"financial_transaction_id": callback.FinancialTransactionID,
		"external_id":              callback.ExternalID,
		"amount":                   callback.Amount,
		"currency":                 callback.Currency,
		"mtn_status":               callback.Status,
		"entity_type":              entityType,
	}

	if callback.Reason != nil {
		extras["reason_code"] = callback.Reason.Code
		extras["reason_message"] = callback.Reason.Message
	}

	statusReq := &commonv1.StatusUpdateRequest{
		Id:         callback.ExternalID,
		State:      state,
		Status:     status,
		ExternalId: callback.FinancialTransactionID,
		Extras:     extras.ToProtoStruct(),
	}

	if _, err := s.paymentCli.StatusUpdate(ctx, connect.NewRequest(statusReq)); err != nil {
		logger.WithError(err).Error("could not update payment status")
		s.metrics.WebhookRejected(ctx, entityType, "status_update_error")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// mapMtnStatus maps MTN MoMo status strings to internal status enum.
func mapMtnStatus(mtnStatus string) commonv1.STATUS {
	switch mtnStatus {
	case "SUCCESSFUL":
		return commonv1.STATUS_SUCCESSFUL
	case "FAILED":
		return commonv1.STATUS_FAILED
	case "PENDING":
		return commonv1.STATUS_IN_PROCESS
	default:
		return commonv1.STATUS_UNKNOWN
	}
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
