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

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/mtn/service/client"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/frame/v2/security"
	"github.com/pitabwire/util"
)

// CredentialsResolver resolves MTN credentials for a settings connection key
// (empty = service defaults). Used to re-query transaction status.
type CredentialsResolver func(ctx context.Context, connection string) (*client.MtnCredentials, error)

// MtnWebhookServer handles MTN MoMo callback webhooks.
//
// MTN callbacks are not signed, so the body is only used to identify the
// entity (externalId). The recorded status, amount and currency always come
// from MTN's status API, queried with the reference id stored when the
// request was issued.
type MtnWebhookServer struct {
	paymentCli   paymentv1connect.PaymentServiceClient
	mtnCli       client.MtnClient
	resolveCreds CredentialsResolver
	metrics      *integrationobs.Metrics
}

// NewMtnWebhookServer creates a new webhook server.
func NewMtnWebhookServer(
	paymentCli paymentv1connect.PaymentServiceClient,
	mtnCli client.MtnClient,
	resolveCreds CredentialsResolver,
) *MtnWebhookServer {
	return &MtnWebhookServer{
		paymentCli:   paymentCli,
		mtnCli:       mtnCli,
		resolveCreds: resolveCreds,
		metrics:      integrationobs.NewMetrics("mtn"),
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

// errNotBound means the callback does not match a request we issued.
var errNotBound = errors.New("callback does not match an issued request")

// confirmed is MTN's authoritative view of a request.
type confirmed struct {
	status, amount, currency, externalID, financialTxID string
	reasonCode, reasonMessage                           string
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

	entityID := strings.TrimSpace(callback.ExternalID)
	logger = logger.WithField("external_id", entityID)
	if entityID == "" {
		s.metrics.WebhookRejected(ctx, entityType, "missing_external_id")
		http.Error(w, "missing externalId", http.StatusBadRequest)
		return
	}

	stored, err := s.loadRequest(ctx, entityID, entityType)
	if err != nil {
		s.reject(ctx, w, logger, entityType, err)
		return
	}

	conf, err := s.query(ctx, entityType, stored)
	if err != nil {
		s.reject(ctx, w, logger, entityType, err)
		return
	}
	if conf.externalID != "" && conf.externalID != entityID {
		s.reject(ctx, w, logger, entityType,
			fmt.Errorf("%w: MTN reports externalId %q", errNotBound, conf.externalID))
		return
	}

	status := mapMtnStatus(conf.status)
	extras := data.JSONMap{
		"financial_transaction_id":        conf.financialTxID,
		"external_id":                     entityID,
		"amount":                          conf.amount,
		"currency":                        conf.currency,
		"mtn_status":                      conf.status,
		"verification":                    "status_query",
		"entity_type":                     entityType,
		client.ExtraReferenceID:           stored.GetString(client.ExtraReferenceID),
		client.ExtraRequestedAmount:       stored.GetString(client.ExtraRequestedAmount),
		client.ExtraRequestedCurrency:     stored.GetString(client.ExtraRequestedCurrency),
		client.ExtraCredentialsConnection: stored.GetString(client.ExtraCredentialsConnection),
	}
	if conf.reasonCode != "" || conf.reasonMessage != "" {
		extras["reason_code"] = conf.reasonCode
		extras["reason_message"] = conf.reasonMessage
	}
	if status == commonv1.STATUS_SUCCESSFUL && !matchesRequest(stored, conf) {
		logger.WithFields(map[string]any{
			"requested_amount": stored.GetString(client.ExtraRequestedAmount),
			"confirmed_amount": conf.amount,
			"confirmed_ccy":    conf.currency,
		}).Error("MTN confirmed a different amount or currency than requested")
		status = commonv1.STATUS_FAILED
		extras["verification"] = "amount_mismatch"
	}

	state := commonv1.STATE_ACTIVE
	if status == commonv1.STATUS_FAILED {
		state = commonv1.STATE_INACTIVE
	}

	statusReq := &commonv1.StatusUpdateRequest{
		Id:         entityID,
		State:      state,
		Status:     status,
		ExternalId: conf.financialTxID,
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
// match an issued request, otherwise 503 so MTN retries.
func (s *MtnWebhookServer) reject(
	ctx context.Context,
	w http.ResponseWriter,
	logger *util.LogEntry,
	entityType string,
	err error,
) {
	if errors.Is(err, errNotBound) {
		logger.WithError(err).Warn("MTN callback rejected")
		s.metrics.WebhookRejected(ctx, entityType, "not_bound")
		http.Error(w, "callback does not match request", http.StatusConflict)
		return
	}
	logger.WithError(err).Error("could not verify MTN callback")
	s.metrics.WebhookRejected(ctx, entityType, "verification_error")
	http.Error(w, "could not verify callback", http.StatusServiceUnavailable)
}

// loadRequest reads the entity's latest status, which carries the reference id
// the request was issued with (the workers set it, and this handler carries it
// forward on every status it writes).
func (s *MtnWebhookServer) loadRequest(ctx context.Context, entityID, entityType string) (data.JSONMap, error) {
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
	if stored.GetString(client.ExtraReferenceID) == "" {
		return nil, fmt.Errorf("%w: no MTN reference id recorded", errNotBound)
	}
	return stored, nil
}

// query asks MTN for the request's current state.
func (s *MtnWebhookServer) query(ctx context.Context, entityType string, stored data.JSONMap) (*confirmed, error) {
	if s.mtnCli == nil || s.resolveCreds == nil {
		return nil, errors.New("MTN status query not configured")
	}
	creds, err := s.resolveCreds(ctx, stored.GetString(client.ExtraCredentialsConnection))
	if err != nil {
		return nil, fmt.Errorf("resolve credentials: %w", err)
	}
	referenceID := stored.GetString(client.ExtraReferenceID)
	if entityType == "payment" {
		st, qErr := s.mtnCli.GetTransferStatus(ctx, creds, referenceID)
		if qErr != nil {
			return nil, fmt.Errorf("transfer status query: %w", qErr)
		}
		c := &confirmed{status: st.Status, amount: st.Amount, currency: st.Currency,
			externalID: st.ExternalID, financialTxID: st.FinancialTransactionID}
		if st.Reason != nil {
			c.reasonCode, c.reasonMessage = st.Reason.Code, st.Reason.Message
		}
		return c, nil
	}
	st, err := s.mtnCli.GetRequestToPayStatus(ctx, creds, referenceID)
	if err != nil {
		return nil, fmt.Errorf("requestToPay status query: %w", err)
	}
	c := &confirmed{status: st.Status, amount: st.Amount, currency: st.Currency,
		externalID: st.ExternalID, financialTxID: st.FinancialTransactionID}
	if st.Reason != nil {
		c.reasonCode, c.reasonMessage = st.Reason.Code, st.Reason.Message
	}
	return c, nil
}

// matchesRequest checks MTN's confirmed amount/currency against what was sent.
// Requests issued before the amount was recorded have nothing to compare.
func matchesRequest(stored data.JSONMap, conf *confirmed) bool {
	if want := stored.GetString(client.ExtraRequestedCurrency); want != "" &&
		!strings.EqualFold(want, conf.currency) {
		return false
	}
	want := stored.GetString(client.ExtraRequestedAmount)
	if want == "" {
		return true
	}
	x, errA := strconv.ParseFloat(strings.TrimSpace(want), 64)
	y, errB := strconv.ParseFloat(strings.TrimSpace(conf.amount), 64)
	return errA == nil && errB == nil && math.Abs(x-y) < amountTolerance
}

// amountTolerance absorbs decimal formatting differences ("100" vs "100.00").
const amountTolerance = 0.005

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
