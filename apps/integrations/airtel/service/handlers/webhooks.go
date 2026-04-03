package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/airtel/service/client"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/security"
	"github.com/pitabwire/util"
)

// AirtelWebhookServer handles Airtel Money callback webhooks.
type AirtelWebhookServer struct {
	paymentCli paymentv1connect.PaymentServiceClient
}

// NewAirtelWebhookServer creates a new webhook server.
func NewAirtelWebhookServer(paymentCli paymentv1connect.PaymentServiceClient) *AirtelWebhookServer {
	return &AirtelWebhookServer{
		paymentCli: paymentCli,
	}
}

// NewRouterV1 creates the HTTP routes for Airtel Money webhooks.
func (s *AirtelWebhookServer) NewRouterV1() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/airtel/collection", s.HandleCollectionCallback)
	mux.HandleFunc("/webhook/airtel/disbursement", s.HandleDisbursementCallback)
	return mux
}

// HandleCollectionCallback processes collection callbacks from Airtel Money.
func (s *AirtelWebhookServer) HandleCollectionCallback(w http.ResponseWriter, r *http.Request) {
	s.handleCallback(w, r, "airtel.webhook.collection", "prompt")
}

// HandleDisbursementCallback processes disbursement callbacks from Airtel Money.
func (s *AirtelWebhookServer) HandleDisbursementCallback(w http.ResponseWriter, r *http.Request) {
	s.handleCallback(w, r, "airtel.webhook.disbursement", "payment")
}

// handleCallback is the shared implementation for collection and disbursement callbacks.
func (s *AirtelWebhookServer) handleCallback(w http.ResponseWriter, r *http.Request, logType, entityType string) {
	ctx := injectTenantFromQuery(r.Context(), r)
	logger := util.Log(ctx).WithField("type", logType)
	defer logger.Release()

	var callback client.CollectionCallbackBody
	if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
		logger.WithError(err).Error("failed to decode callback")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	txn := callback.Transaction
	logger = logger.WithField("transaction_id", txn.ID)

	status := mapAirtelStatus(txn.StatusCode)
	state := commonv1.STATE_ACTIVE
	if status == commonv1.STATUS_FAILED {
		state = commonv1.STATE_INACTIVE
	}

	extras := data.JSONMap{
		"transaction_id":  txn.ID,
		"message":         txn.Message,
		"status_code":     txn.StatusCode,
		"airtel_money_id": txn.AirtelMoney.ID,
		"entity_type":     entityType,
	}

	statusReq := &commonv1.StatusUpdateRequest{
		Id:         txn.ID,
		State:      state,
		Status:     status,
		ExternalId: txn.AirtelMoney.ID,
		Extras:     extras.ToProtoStruct(),
	}

	if _, err := s.paymentCli.StatusUpdate(ctx, connect.NewRequest(statusReq)); err != nil {
		logger.WithError(err).Error("could not update payment status")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// mapAirtelStatus maps Airtel status codes to internal status enum.
func mapAirtelStatus(statusCode string) commonv1.STATUS {
	switch statusCode {
	case "TS":
		return commonv1.STATUS_SUCCESSFUL
	case "TIP":
		return commonv1.STATUS_IN_PROCESS
	case "TF":
		return commonv1.STATUS_FAILED
	case "TA":
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
