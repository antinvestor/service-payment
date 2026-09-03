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
	"io"
	"net/http"
	"strconv"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/service/credentials"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/frame/v2/security"
	"github.com/pitabwire/util"
)

const (
	maxWebhookBody = 1 << 20 // 1 MiB

	kindReceive = "receive"
	kindSend    = "send"

	providerName = "yellowcard"
)

// YellowcardWebhookServer handles Yellow Card event webhooks.
//
// A webhook is first authenticated with the X-YC-Signature HMAC. Even then
// the body is only used to learn which record to look at: the authoritative
// status and amounts are re-fetched from the Yellow Card API before any
// payment status is updated. Yellow Card records carry no partner metadata,
// so tenancy comes from the tenant_id / partition_id query parameters the
// operator configures on the webhook URL.
type YellowcardWebhookServer struct {
	paymentCli    paymentv1connect.PaymentServiceClient
	ycCli         client.YellowcardClient
	credsResolver *credentials.Resolver
	metrics       *integrationobs.Metrics
}

// NewYellowcardWebhookServer creates a new webhook server.
func NewYellowcardWebhookServer(
	paymentCli paymentv1connect.PaymentServiceClient,
	ycCli client.YellowcardClient,
	credsResolver *credentials.Resolver,
) *YellowcardWebhookServer {
	return &YellowcardWebhookServer{
		paymentCli:    paymentCli,
		ycCli:         ycCli,
		credsResolver: credsResolver,
		metrics:       integrationobs.NewMetrics(providerName),
	}
}

// NewRouterV1 creates the HTTP routes for Yellow Card webhooks.
func (s *YellowcardWebhookServer) NewRouterV1() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/yellowcard/receives", func(w http.ResponseWriter, r *http.Request) {
		s.handle(w, r, kindReceive)
	})
	mux.HandleFunc("/webhook/yellowcard/sends", func(w http.ResponseWriter, r *http.Request) {
		s.handle(w, r, kindSend)
	})
	mux.HandleFunc("/webhook/yellowcard", func(w http.ResponseWriter, r *http.Request) {
		s.handle(w, r, "")
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

// webhookEvent is the Yellow Card webhook payload.
type webhookEvent struct {
	ID         string `json:"id"`
	SequenceID string `json:"sequenceId"`
	Status     string `json:"status"`
	APIKey     string `json:"apiKey"`
	Event      string `json:"event"`
	ErrorCode  string `json:"errorCode"`
	SessionID  string `json:"sessionId"`
	ExecutedAt string `json:"executedAt"`
}

// verified is the provider-agnostic content of a re-fetched record.
type verified struct {
	kind            string
	id              string
	sequenceID      string
	status          string
	errorCode       string
	country         string
	currency        string
	convertedAmount float64
	rate            float64
	reference       string
	bankInfo        *client.BankInfo
}

// kindFromEvent maps a webhook event name to a record kind, accepting both
// the v2 (RECEIVE./SEND.) and legacy (COLLECTION./PAYMENT.) prefixes.
func kindFromEvent(event string) string {
	upper := strings.ToUpper(event)
	switch {
	case strings.HasPrefix(upper, "RECEIVE."), strings.HasPrefix(upper, "COLLECTION."):
		return kindReceive
	case strings.HasPrefix(upper, "SEND."), strings.HasPrefix(upper, "PAYMENT."):
		return kindSend
	}
	return ""
}

func (s *YellowcardWebhookServer) handle(w http.ResponseWriter, r *http.Request, kind string) {
	ctx := r.Context()
	metricKind := kind
	if metricKind == "" {
		metricKind = "event"
	}
	s.metrics.WebhookReceived(ctx, metricKind)
	logger := util.Log(ctx).WithField("type", "yellowcard.webhook."+metricKind)
	defer logger.Release()

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxWebhookBody))
	if err != nil {
		s.metrics.WebhookRejected(ctx, metricKind, "decode_error")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var event webhookEvent
	if err = json.Unmarshal(body, &event); err != nil {
		logger.WithError(err).Error("failed to decode webhook")
		s.metrics.WebhookRejected(ctx, metricKind, "decode_error")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if event.ID == "" && event.SequenceID == "" {
		s.metrics.WebhookRejected(ctx, metricKind, "missing_id")
		http.Error(w, "missing payment id", http.StatusBadRequest)
		return
	}

	creds, err := s.resolveCredentials(ctx, r.URL.Query().Get("connection"))
	if err != nil {
		logger.WithError(err).Error("cannot resolve credentials to verify webhook, rejecting")
		s.metrics.WebhookRejected(ctx, metricKind, "verification_failed")
		http.Error(w, "webhook verification unavailable", http.StatusServiceUnavailable)
		return
	}

	if !client.VerifyWebhookSignature(body, r.Header.Get(client.HeaderWebhookSignature), creds.ResolveWebhookSecret()) {
		logger.Error("webhook signature verification failed")
		s.metrics.WebhookRejected(ctx, metricKind, "verification_failed")
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}
	if event.APIKey != "" && event.APIKey != creds.APIKey {
		logger.Error("webhook api key does not match the verifying credentials")
		s.metrics.WebhookRejected(ctx, metricKind, "verification_failed")
		http.Error(w, "invalid api key", http.StatusUnauthorized)
		return
	}

	if kind == "" {
		kind = kindFromEvent(event.Event)
	}
	if kind == "" {
		s.metrics.WebhookRejected(ctx, metricKind, "decode_error")
		http.Error(w, "unknown event", http.StatusBadRequest)
		return
	}

	record, err := s.fetch(ctx, creds, kind, &event)
	if client.IsNotFound(err) {
		logger.WithField("sequence_id", event.SequenceID).Error("webhook for record unknown to Yellow Card, rejecting")
		s.metrics.WebhookRejected(ctx, kind, "unknown_payment")
		http.Error(w, "unknown payment", http.StatusNotFound)
		return
	}
	if err != nil {
		logger.WithError(err).Error("could not verify webhook against Yellow Card")
		s.metrics.WebhookRejected(ctx, kind, "verification_failed")
		http.Error(w, "could not verify webhook", http.StatusBadGateway)
		return
	}

	s.process(w, r, record, &event)
}

func (s *YellowcardWebhookServer) resolveCredentials(ctx context.Context, connection string) (*client.Credentials, error) {
	if connection != "" {
		creds, err := s.credsResolver.FromConnection(ctx, connection)
		if err == nil {
			return creds, nil
		}
		util.Log(ctx).WithError(err).Warn("connection lookup failed, falling back to default credentials")
	}
	return s.credsResolver.Default()
}

func (s *YellowcardWebhookServer) fetch(
	ctx context.Context,
	creds *client.Credentials,
	kind string,
	event *webhookEvent,
) (*verified, error) {
	if kind == kindSend {
		var (
			snd *client.Send
			err error
		)
		if event.SequenceID != "" {
			snd, err = s.ycCli.GetSendBySequenceID(ctx, creds, event.SequenceID)
		} else {
			snd, err = s.ycCli.GetSend(ctx, creds, event.ID)
		}
		if err != nil {
			return nil, err
		}
		return &verified{
			kind: kindSend, id: snd.ID, sequenceID: snd.SequenceID, status: snd.Status, errorCode: snd.ErrorCode,
			country: snd.Country, currency: snd.Currency, convertedAmount: snd.ConvertedAmount, rate: snd.Rate,
			reference: snd.Reference,
		}, nil
	}

	var (
		rcv *client.Receive
		err error
	)
	if event.SequenceID != "" {
		rcv, err = s.ycCli.GetReceiveBySequenceID(ctx, creds, event.SequenceID)
	} else {
		rcv, err = s.ycCli.GetReceive(ctx, creds, event.ID)
	}
	if err != nil {
		return nil, err
	}
	return &verified{
		kind: kindReceive, id: rcv.ID, sequenceID: rcv.SequenceID, status: rcv.Status, errorCode: rcv.ErrorCode,
		country: rcv.Country, currency: rcv.Currency, convertedAmount: rcv.ConvertedAmount, rate: rcv.Rate,
		reference: rcv.Reference, bankInfo: rcv.BankInfo,
	}, nil
}

func (s *YellowcardWebhookServer) process(w http.ResponseWriter, r *http.Request, v *verified, event *webhookEvent) {
	logger := util.Log(r.Context()).
		WithField("type", "yellowcard.webhook."+v.kind).
		WithField(v.kind+"_id", v.id).
		WithField("sequence_id", v.sequenceID)
	defer logger.Release()

	ctx := injectTenant(r)

	entityType := "prompt"
	if v.kind == kindSend {
		entityType = "payment"
	}
	if v.sequenceID == "" {
		logger.Error("verified record carries no sequence id, cannot correlate")
		s.metrics.WebhookRejected(r.Context(), v.kind, "unknown_payment")
		http.Error(w, "record has no sequence id", http.StatusNotFound)
		return
	}

	status, state := MapStatus(v.status)
	extras := data.JSONMap{
		"entity_type":             entityType,
		collection.ExtraProvider:  providerName,
		v.kind + "_id":            v.id,
		"sequence_id":             v.sequenceID,
		"status":                  v.status,
		"event":                   event.Event,
		"executed_at":             event.ExecutedAt,
		"country":                 v.country,
		"currency":                v.currency,
		"local_amount":            strconv.FormatFloat(v.convertedAmount, 'f', -1, 64),
		"rate":                    strconv.FormatFloat(v.rate, 'f', -1, 64),
		"provider_transaction_id": v.reference,
	}
	if code := firstNonEmpty(v.errorCode, event.ErrorCode); code != "" || status == commonv1.STATUS_FAILED {
		if code == "" {
			code = strings.ToUpper(v.status)
		}
		extras["error_code"] = code
		extras["failure_code"] = code
		extras["failure_message"] = failureMessage(code, v.status)
	}
	if isRefundStatus(v.status) {
		extras["refund_status"] = v.status
	}
	if v.kind == kindReceive && v.bankInfo != nil {
		if v.bankInfo.Name != "" {
			extras[collection.ExtraBankName] = v.bankInfo.Name
		}
		if v.bankInfo.AccountNumber != "" {
			extras[collection.ExtraBankAccountNumber] = v.bankInfo.AccountNumber
		}
		if v.bankInfo.AccountName != "" {
			extras[collection.ExtraBankAccountName] = v.bankInfo.AccountName
		}
		if v.reference != "" {
			extras[collection.ExtraPaymentReference] = v.reference
		}
	}

	statusReq := &commonv1.StatusUpdateRequest{
		Id:         v.sequenceID,
		State:      state,
		Status:     status,
		ExternalId: v.id,
		Extras:     extras.ToProtoStruct(),
	}

	if _, err := s.paymentCli.StatusUpdate(ctx, connect.NewRequest(statusReq)); err != nil {
		logger.WithError(err).Error("could not update payment status")
		s.metrics.WebhookRejected(r.Context(), v.kind, "status_update_error")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// MapStatus maps a Yellow Card status to the platform status and state.
func MapStatus(status string) (commonv1.STATUS, commonv1.STATE) {
	switch strings.ToLower(status) {
	case client.StatusComplete:
		return commonv1.STATUS_SUCCESSFUL, commonv1.STATE_ACTIVE
	case client.StatusFailed, client.StatusExpired, client.StatusCancelled, client.StatusRefunded:
		return commonv1.STATUS_FAILED, commonv1.STATE_INACTIVE
	case "":
		return commonv1.STATUS_UNKNOWN, commonv1.STATE_ACTIVE
	default:
		return commonv1.STATUS_IN_PROCESS, commonv1.STATE_ACTIVE
	}
}

func isRefundStatus(status string) bool {
	switch strings.ToLower(status) {
	case client.StatusPendingRefund, client.StatusRefundProcessing, client.StatusRefundFailed, client.StatusRefunded:
		return true
	}
	return false
}

var failureMessages = map[string]string{
	client.TxErrExpired:             "The payment was not completed before it expired",
	client.TxErrInvalidRecipient:    "The recipient details failed validation",
	client.TxErrValidationFailed:    "The transaction failed validation",
	client.TxErrInvalidNetwork:      "The selected mobile money or bank network is invalid",
	client.TxErrInvalidCurrency:     "The currency is not supported for this country",
	client.TxErrInsufficientBalance: "Insufficient balance to complete the transaction",
	client.TxErrRefused:             "The customer did not approve the payment",
	client.TxErrGatewayTimeout:      "The payment provider timed out",
	client.TxErrProviderError:       "The payment provider returned an error",
	client.TxErrPossibleDuplicate:   "The provider flagged the transaction as a possible duplicate",
	client.TxErrNameMismatch:        "The payer name does not match the customer on record",
	client.TxErrFraudCheck:          "The transaction was declined by risk checks",
	client.TxErrOtherError:          "The transaction failed",
	"CANCELLED":                     "The payment was cancelled",
	"REFUNDED":                      "The payment was refunded to the customer",
}

func failureMessage(code, status string) string {
	if msg, ok := failureMessages[strings.ToUpper(code)]; ok {
		return msg
	}
	return "The payment " + strings.ToLower(status)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// injectTenant builds the claims context from the operator-configured
// webhook URL query string. Yellow Card records carry no partner metadata,
// so the query string is the only tenancy signal; the webhook has already
// been authenticated by signature before this point.
func injectTenant(r *http.Request) context.Context {
	ctx := r.Context()
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
