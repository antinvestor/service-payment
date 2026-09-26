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
	"net/http"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/airtel/service/client"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/frame/v2/security"
	"github.com/pitabwire/util"
)

// CredentialsResolver resolves Airtel credentials for a settings connection
// key (empty = service defaults). Used to re-query transaction status.
type CredentialsResolver func(ctx context.Context, connection string) (*client.AirtelCredentials, error)

// AirtelWebhookServer handles Airtel Money callback webhooks.
//
// Airtel callbacks are not signed, so the body only identifies our
// transaction id; the recorded status comes from Airtel's enquiry API.
type AirtelWebhookServer struct {
	paymentCli   paymentv1connect.PaymentServiceClient
	airtelCli    client.AirtelClient
	resolveCreds CredentialsResolver
	metrics      *integrationobs.Metrics
}

// NewAirtelWebhookServer creates a new webhook server.
func NewAirtelWebhookServer(
	paymentCli paymentv1connect.PaymentServiceClient,
	airtelCli client.AirtelClient,
	resolveCreds CredentialsResolver,
) *AirtelWebhookServer {
	return &AirtelWebhookServer{
		paymentCli:   paymentCli,
		airtelCli:    airtelCli,
		resolveCreds: resolveCreds,
		metrics:      integrationobs.NewMetrics("airtel"),
	}
}

// NewRouterV1 creates the HTTP routes for Airtel Money webhooks.
func (s *AirtelWebhookServer) NewRouterV1() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/airtel/collection", s.HandleCollectionCallback)
	mux.HandleFunc("/webhook/airtel/disbursement", s.HandleDisbursementCallback)
	return mux
}

// HandleCollectionCallback processes collection (USSD push) callbacks.
func (s *AirtelWebhookServer) HandleCollectionCallback(w http.ResponseWriter, r *http.Request) {
	s.metrics.WebhookReceived(r.Context(), "collection")
	s.handleCallback(w, r, "airtel.webhook.collection", "prompt")
}

// HandleDisbursementCallback processes disbursement callbacks.
func (s *AirtelWebhookServer) HandleDisbursementCallback(w http.ResponseWriter, r *http.Request) {
	s.metrics.WebhookReceived(r.Context(), "disbursement")
	s.handleCallback(w, r, "airtel.webhook.disbursement", "payment")
}

// errNotBound means the callback does not match a request we issued.
var errNotBound = errors.New("callback does not match an issued request")

// handleCallback is the shared logic for processing Airtel callbacks.
func (s *AirtelWebhookServer) handleCallback(w http.ResponseWriter, r *http.Request, logType, entityType string) {
	ctx := injectTenantFromQuery(r.Context(), r)
	logger := util.Log(ctx).WithField("type", logType)
	defer logger.Release()

	var callback client.CollectionCallbackBody
	if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
		logger.WithError(err).Error("failed to decode callback")
		s.metrics.WebhookRejected(ctx, entityType, "decode_error")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	entityID := strings.TrimSpace(callback.Transaction.ID)
	logger = logger.WithField("transaction_id", entityID)
	if entityID == "" {
		s.metrics.WebhookRejected(ctx, entityType, "missing_transaction_id")
		http.Error(w, "missing transaction id", http.StatusBadRequest)
		return
	}

	stored, err := s.loadRequest(ctx, entityID, entityType)
	if err != nil {
		s.reject(ctx, w, logger, entityType, err)
		return
	}

	txn, err := s.query(ctx, entityType, entityID, stored)
	if err != nil {
		s.reject(ctx, w, logger, entityType, err)
		return
	}

	status := mapAirtelStatus(txn.Status)
	state := commonv1.STATE_ACTIVE
	if status == commonv1.STATUS_FAILED {
		state = commonv1.STATE_INACTIVE
	}

	extras := data.JSONMap{
		"transaction_id":                  entityID,
		"message":                         txn.Message,
		"status_code":                     txn.Status,
		"airtel_money_id":                 txn.AirtelMoneyID,
		"verification":                    "status_query",
		"entity_type":                     entityType,
		client.ExtraRequestedAmount:       stored.GetString(client.ExtraRequestedAmount),
		client.ExtraRequestedCurrency:     stored.GetString(client.ExtraRequestedCurrency),
		client.ExtraCredentialsConnection: stored.GetString(client.ExtraCredentialsConnection),
	}

	statusReq := &commonv1.StatusUpdateRequest{
		Id:         entityID,
		State:      state,
		Status:     status,
		ExternalId: txn.AirtelMoneyID,
		Extras:     extras.ToProtoStruct(),
	}

	if _, err = s.paymentCli.StatusUpdate(ctx, connect.NewRequest(statusReq)); err != nil {
		logger.WithError(err).Error("could not update payment status")
		s.metrics.WebhookRejected(ctx, entityType, "status_update_error")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// reject answers without recording anything: 409 when the callback does not
// match an issued request, otherwise 503 so Airtel retries.
func (s *AirtelWebhookServer) reject(
	ctx context.Context,
	w http.ResponseWriter,
	logger *util.LogEntry,
	entityType string,
	err error,
) {
	if errors.Is(err, errNotBound) {
		logger.WithError(err).Warn("Airtel callback rejected")
		s.metrics.WebhookRejected(ctx, entityType, "not_bound")
		http.Error(w, "callback does not match request", http.StatusConflict)
		return
	}
	logger.WithError(err).Error("could not verify Airtel callback")
	s.metrics.WebhookRejected(ctx, entityType, "verification_error")
	http.Error(w, "could not verify callback", http.StatusServiceUnavailable)
}

// loadRequest reads the entity's latest status; the workers record Airtel's
// transaction_id there once the request was accepted, and this handler
// carries the binding forward on every status it writes.
func (s *AirtelWebhookServer) loadRequest(ctx context.Context, entityID, entityType string) (data.JSONMap, error) {
	reqExtras := data.JSONMap{"entity_type": entityType}
	resp, err := s.paymentCli.Status(ctx, connect.NewRequest(&commonv1.StatusRequest{
		Id:     entityID,
		Extras: reqExtras.ToProtoStruct(),
	}))
	if err != nil {
		return nil, fmt.Errorf("load %s status: %w", entityType, err)
	}
	var stored data.JSONMap
	stored = stored.FromProtoStruct(resp.Msg.GetExtras())
	if _, ok := stored[client.ExtraTransactionID]; !ok {
		return nil, fmt.Errorf("%w: no Airtel request recorded", errNotBound)
	}
	return stored, nil
}

// airtelTxn is Airtel's authoritative view of a transaction.
type airtelTxn struct {
	Status, Message, AirtelMoneyID string
}

// query asks Airtel for the transaction's current state.
func (s *AirtelWebhookServer) query(
	ctx context.Context,
	entityType, entityID string,
	stored data.JSONMap,
) (*airtelTxn, error) {
	if s.airtelCli == nil || s.resolveCreds == nil {
		return nil, errors.New("airtel status query not configured")
	}
	creds, err := s.resolveCreds(ctx, stored.GetString(client.ExtraCredentialsConnection))
	if err != nil {
		return nil, fmt.Errorf("resolve credentials: %w", err)
	}
	if ccy := stored.GetString(client.ExtraRequestedCurrency); ccy != "" {
		creds.Currency = ccy
	}
	var resp *client.StatusResponse
	if entityType == "payment" {
		resp, err = s.airtelCli.DisbursementStatus(ctx, creds, entityID)
	} else {
		resp, err = s.airtelCli.TransactionStatus(ctx, creds, entityID)
	}
	if err != nil {
		return nil, fmt.Errorf("airtel status query: %w", err)
	}
	t := resp.Data.Transaction
	if t.ID != "" && t.ID != entityID {
		return nil, fmt.Errorf("%w: Airtel reports transaction %q", errNotBound, t.ID)
	}
	if t.Status == "" {
		return nil, errors.New("airtel status query returned no transaction status")
	}
	return &airtelTxn{Status: t.Status, Message: t.Message, AirtelMoneyID: t.AirtelMoneyID}, nil
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
