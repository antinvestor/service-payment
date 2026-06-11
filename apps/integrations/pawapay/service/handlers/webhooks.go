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

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/credentials"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/security"
	"github.com/pitabwire/util"
)

// PawapayWebhookServer handles pawaPay callback webhooks.
//
// Callbacks are treated as untrusted notifications: the request body is only
// used to learn which payment to look at. The authoritative status, amounts
// and metadata are re-fetched from the pawaPay API with our credentials
// before any payment status is updated, so a forged callback cannot move a
// payment to a state pawaPay does not confirm. Tenancy is likewise recovered
// from the verified record's metadata (attached at initiation), never from
// the unauthenticated request body.
type PawapayWebhookServer struct {
	paymentCli    paymentv1connect.PaymentServiceClient
	pawapayCli    client.PawapayClient
	credsResolver *credentials.Resolver
}

// NewPawapayWebhookServer creates a new webhook server.
func NewPawapayWebhookServer(
	paymentCli paymentv1connect.PaymentServiceClient,
	pawapayCli client.PawapayClient,
	credsResolver *credentials.Resolver,
) *PawapayWebhookServer {
	return &PawapayWebhookServer{
		paymentCli:    paymentCli,
		pawapayCli:    pawapayCli,
		credsResolver: credsResolver,
	}
}

// NewRouterV1 creates the HTTP routes for pawaPay webhooks.
func (s *PawapayWebhookServer) NewRouterV1() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/pawapay/deposits", s.HandleDepositCallback)
	mux.HandleFunc("/webhook/pawapay/payouts", s.HandlePayoutCallback)
	mux.HandleFunc("/webhook/pawapay/refunds", s.HandleRefundCallback)
	return mux
}

// callbackData is the provider-type-agnostic content of a verified payment record.
type callbackData struct {
	kind                  string // deposit, payout or refund
	id                    string
	status                string
	amount                string
	currency              string
	country               string
	providerTransactionID string
	party                 *client.PayerOrRecipient
	failureReason         *client.FailureReason
	metadata              map[string]any
	defaultEntityType     string
}

// callbackNotification is the only information taken from the untrusted
// callback body: which payment to verify and an optional hint about which
// credential connection it belongs to.
type callbackNotification struct {
	DepositID string         `json:"depositId"`
	PayoutID  string         `json:"payoutId"`
	RefundID  string         `json:"refundId"`
	Metadata  map[string]any `json:"metadata"`
}

// errPaymentNotFound indicates the payment a callback refers to is unknown to pawaPay.
var errPaymentNotFound = errors.New("payment not found in pawaPay")

// verifiedFetch looks up a payment on the pawaPay API and maps it to the
// generic callbackData. It returns errPaymentNotFound when the payment is
// unknown to pawaPay.
type verifiedFetch func(ctx context.Context, creds *client.Credentials, id string) (*callbackData, error)

// HandleDepositCallback processes final deposit callbacks from pawaPay.
func (s *PawapayWebhookServer) HandleDepositCallback(w http.ResponseWriter, r *http.Request) {
	s.handleCallback(w, r, "deposit",
		func(ctx context.Context, creds *client.Credentials, id string) (*callbackData, error) {
			result, err := s.pawapayCli.GetDeposit(ctx, creds, id)
			if err != nil {
				return nil, err
			}
			if result.Status != client.SearchStatusFound || result.Data == nil {
				return nil, errPaymentNotFound
			}
			verified := result.Data
			cb := newCallbackData(
				"deposit",
				"prompt",
				verified.DepositID,
				verified.Status,
				verified.Payer,
				verified.Metadata,
			)
			cb.amount, cb.currency, cb.country = verified.Amount, verified.Currency, verified.Country
			cb.providerTransactionID, cb.failureReason = verified.ProviderTransactionID, verified.FailureReason
			return cb, nil
		})
}

// HandlePayoutCallback processes final payout callbacks from pawaPay.
func (s *PawapayWebhookServer) HandlePayoutCallback(w http.ResponseWriter, r *http.Request) {
	s.handleCallback(w, r, "payout",
		func(ctx context.Context, creds *client.Credentials, id string) (*callbackData, error) {
			result, err := s.pawapayCli.GetPayout(ctx, creds, id)
			if err != nil {
				return nil, err
			}
			if result.Status != client.SearchStatusFound || result.Data == nil {
				return nil, errPaymentNotFound
			}
			verified := result.Data
			cb := newCallbackData(
				"payout",
				"payment",
				verified.PayoutID,
				verified.Status,
				verified.Recipient,
				verified.Metadata,
			)
			cb.amount, cb.currency, cb.country = verified.Amount, verified.Currency, verified.Country
			cb.providerTransactionID, cb.failureReason = verified.ProviderTransactionID, verified.FailureReason
			return cb, nil
		})
}

// HandleRefundCallback processes final refund callbacks from pawaPay.
func (s *PawapayWebhookServer) HandleRefundCallback(w http.ResponseWriter, r *http.Request) {
	s.handleCallback(w, r, "refund",
		func(ctx context.Context, creds *client.Credentials, id string) (*callbackData, error) {
			result, err := s.pawapayCli.GetRefund(ctx, creds, id)
			if err != nil {
				return nil, err
			}
			if result.Status != client.SearchStatusFound || result.Data == nil {
				return nil, errPaymentNotFound
			}
			verified := result.Data
			cb := newCallbackData(
				"refund",
				"payment",
				verified.RefundID,
				verified.Status,
				verified.Recipient,
				verified.Metadata,
			)
			cb.amount, cb.currency, cb.country = verified.Amount, verified.Currency, verified.Country
			cb.providerTransactionID, cb.failureReason = verified.ProviderTransactionID, verified.FailureReason
			return cb, nil
		})
}

func newCallbackData(
	kind, defaultEntityType, id, status string,
	party *client.PayerOrRecipient,
	metadata map[string]any,
) *callbackData {
	return &callbackData{
		kind:              kind,
		defaultEntityType: defaultEntityType,
		id:                id,
		status:            status,
		party:             party,
		metadata:          metadata,
	}
}

// handleCallback runs the shared webhook flow: decode the untrusted
// notification, resolve credentials, verify the payment against pawaPay and
// emit a status update built from the verified record only.
func (s *PawapayWebhookServer) handleCallback(
	w http.ResponseWriter,
	r *http.Request,
	kind string,
	fetch verifiedFetch,
) {
	notification, creds, ok := s.acceptNotification(w, r, kind)
	if !ok {
		return
	}

	id := notificationID(notification, kind)
	cb, err := fetch(r.Context(), creds, id)
	if errors.Is(err, errPaymentNotFound) {
		s.rejectUnknown(w, r, kind, id)
		return
	}
	if err != nil {
		s.rejectUnverifiable(w, r, kind, err)
		return
	}

	s.processCallback(w, r, cb)
}

// acceptNotification decodes the untrusted callback body and resolves the
// credentials needed to verify it against the pawaPay API. It fails closed:
// without resolvable credentials no callback is processed.
func (s *PawapayWebhookServer) acceptNotification(
	w http.ResponseWriter,
	r *http.Request,
	kind string,
) (*callbackNotification, *client.Credentials, bool) {
	logger := util.Log(r.Context()).WithField("type", "pawapay.webhook."+kind)
	defer logger.Release()

	var notification callbackNotification
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		logger.WithError(err).Error("failed to decode callback")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return nil, nil, false
	}

	if notificationID(&notification, kind) == "" {
		logger.Error("callback missing payment id")
		http.Error(w, "missing payment id", http.StatusBadRequest)
		return nil, nil, false
	}

	creds, err := s.resolveCredentials(r.Context(), notification.Metadata)
	if err != nil {
		logger.WithError(err).Error("cannot resolve credentials to verify callback, rejecting")
		http.Error(w, "callback verification unavailable", http.StatusServiceUnavailable)
		return nil, nil, false
	}

	return &notification, creds, true
}

// resolveCredentials picks the credentials used to verify a callback. The
// connection hint from the body only selects which of our stored credential
// sets to use; verification itself happens against the pawaPay API.
func (s *PawapayWebhookServer) resolveCredentials(
	ctx context.Context,
	metadata map[string]any,
) (*client.Credentials, error) {
	if connectionHint := metadataString(metadata, "connection", ""); connectionHint != "" {
		creds, err := s.credsResolver.FromConnection(ctx, connectionHint)
		if err == nil {
			return creds, nil
		}
		util.Log(ctx).WithError(err).
			Warn("connection hint lookup failed, falling back to default credentials")
	}
	return s.credsResolver.Default()
}

func notificationID(notification *callbackNotification, kind string) string {
	switch kind {
	case "deposit":
		return notification.DepositID
	case "payout":
		return notification.PayoutID
	default:
		return notification.RefundID
	}
}

func (s *PawapayWebhookServer) rejectUnverifiable(
	w http.ResponseWriter, r *http.Request, kind string, err error,
) {
	logger := util.Log(r.Context()).WithField("type", "pawapay.webhook."+kind)
	defer logger.Release()
	logger.WithError(err).Error("could not verify callback against pawaPay")
	http.Error(w, "could not verify callback", http.StatusBadGateway)
}

func (s *PawapayWebhookServer) rejectUnknown(
	w http.ResponseWriter, r *http.Request, kind, id string,
) {
	logger := util.Log(r.Context()).WithField("type", "pawapay.webhook."+kind).WithField("id", id)
	defer logger.Release()
	logger.Error("callback for payment unknown to pawaPay, rejecting")
	http.Error(w, "unknown payment", http.StatusNotFound)
}

// processCallback translates a verified pawaPay payment record into a payment
// status update.
func (s *PawapayWebhookServer) processCallback(w http.ResponseWriter, r *http.Request, cb *callbackData) {
	logger := util.Log(r.Context()).
		WithField("type", "pawapay.webhook."+cb.kind).
		WithField(cb.kind+"_id", cb.id)
	defer logger.Release()

	ctx, ok := injectTenant(r, cb.metadata)
	if !ok {
		logger.Error("callback query tenant conflicts with verified payment tenant, rejecting")
		http.Error(w, "tenant mismatch", http.StatusForbidden)
		return
	}

	extras := data.JSONMap{
		cb.kind + "_id":           cb.id,
		"status":                  cb.status,
		"amount":                  cb.amount,
		"currency":                cb.currency,
		"country":                 cb.country,
		"provider_transaction_id": cb.providerTransactionID,
		"entity_type":             metadataString(cb.metadata, "entityType", cb.defaultEntityType),
	}
	if cb.party != nil {
		extras["phone_number"] = cb.party.AccountDetails.PhoneNumber
		extras["provider"] = cb.party.AccountDetails.Provider
	}
	if cb.failureReason != nil {
		extras["failure_code"] = cb.failureReason.FailureCode
		extras["failure_message"] = cb.failureReason.FailureMessage
	}

	statusReq := &commonv1.StatusUpdateRequest{
		Id:         metadataString(cb.metadata, "paymentId", cb.id),
		State:      callbackState(cb.status),
		Status:     callbackStatus(cb.status),
		ExternalId: cb.id,
		Extras:     extras.ToProtoStruct(),
	}

	if _, err := s.paymentCli.StatusUpdate(ctx, connect.NewRequest(statusReq)); err != nil {
		logger.WithError(err).Error("could not update payment status")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// callbackStatus maps a pawaPay payment status to the platform payment status.
func callbackStatus(status string) commonv1.STATUS {
	switch status {
	case client.PaymentStatusCompleted:
		return commonv1.STATUS_SUCCESSFUL
	case client.PaymentStatusFailed:
		return commonv1.STATUS_FAILED
	case client.PaymentStatusAccepted, client.PaymentStatusEnqueued,
		client.PaymentStatusProcessing, client.PaymentStatusInReconciliation:
		return commonv1.STATUS_IN_PROCESS
	default:
		return commonv1.STATUS_UNKNOWN
	}
}

// callbackState maps a pawaPay payment status to the platform record state.
func callbackState(status string) commonv1.STATE {
	if status == client.PaymentStatusFailed {
		return commonv1.STATE_INACTIVE
	}
	return commonv1.STATE_ACTIVE
}

// metadataString reads a string value from callback metadata, returning
// fallback when the key is absent or empty.
func metadataString(metadata map[string]any, key, fallback string) string {
	v, ok := metadata[key]
	if !ok {
		return fallback
	}
	if s, isString := v.(string); isString && s != "" {
		return s
	}
	if s := fmt.Sprintf("%v", v); s != "" && s != "<nil>" {
		return s
	}
	return fallback
}

// injectTenant recovers tenancy for a callback. The verified payment
// record's metadata (attached at initiation and re-fetched from pawaPay) is
// authoritative; query params on the dashboard-configured callback URL only
// fill in tenancy for payments that carry no metadata, because the query
// string is attacker-controlled on an unauthenticated endpoint. When both
// are present and disagree, the callback is rejected (ok=false).
func injectTenant(r *http.Request, verifiedMetadata map[string]any) (context.Context, bool) {
	ctx := r.Context()

	tenantID := metadataString(verifiedMetadata, "tenantId", "")
	partitionID := metadataString(verifiedMetadata, "partitionId", "")

	queryTenantID := r.URL.Query().Get("tenant_id")
	queryPartitionID := r.URL.Query().Get("partition_id")
	if (tenantID != "" && queryTenantID != "" && queryTenantID != tenantID) ||
		(partitionID != "" && queryPartitionID != "" && queryPartitionID != partitionID) {
		return ctx, false
	}
	if tenantID == "" {
		tenantID = queryTenantID
	}
	if partitionID == "" {
		partitionID = queryPartitionID
	}

	if tenantID == "" && partitionID == "" {
		return ctx, true
	}

	claims := &security.AuthenticationClaims{
		TenantID:    tenantID,
		PartitionID: partitionID,
	}
	return claims.ClaimsToContext(ctx), true
}
