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
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/credentials"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	"github.com/pitabwire/frame/v3/data"
	"github.com/pitabwire/frame/v3/security"
	"github.com/pitabwire/util"
)

const (
	maxWebhookBody = 1 << 20 // 1 MiB

	kindCollection = "collection"
	kindPayout     = "payout"

	providerName = "fincra"

	// fincraReferencePrefix marks references Fincra generates (pay codes,
	// charge references); ours are prompt and payment ids.
	fincraReferencePrefix = "fcr-"
)

// FincraWebhookServer handles Fincra webhooks.
//
// A webhook must carry a valid HMAC-SHA512 signature. Even then its body
// only says which record to look at: status, amounts and tenancy are read
// back from the Fincra API by our own reference before any status is
// recorded. Hosted checkouts carry tenancy in the metadata stored on
// Fincra's record; direct charges carry it in their per-request webhook URL.
// Payouts cannot carry it (Fincra's payout API takes no webhook URL or
// metadata), so they take the tenant_id / partition_id query parameters of
// the account's dashboard webhook URL: one Fincra account per paying
// tenant/partition.
type FincraWebhookServer struct {
	paymentCli    paymentv1connect.PaymentServiceClient
	cli           client.FincraClient
	credsResolver *credentials.Resolver
	metrics       *integrationobs.Metrics
}

// NewFincraWebhookServer creates a new webhook server.
func NewFincraWebhookServer(
	paymentCli paymentv1connect.PaymentServiceClient,
	cli client.FincraClient,
	credsResolver *credentials.Resolver,
) *FincraWebhookServer {
	return &FincraWebhookServer{
		paymentCli:    paymentCli,
		cli:           cli,
		credsResolver: credsResolver,
		metrics:       integrationobs.NewMetrics(providerName),
	}
}

// NewRouterV1 creates the HTTP routes for Fincra webhooks.
func (s *FincraWebhookServer) NewRouterV1() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/fincra", s.handle)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

// webhookEvent is the Fincra webhook envelope.
type webhookEvent struct {
	Event string      `json:"event"`
	Data  webhookData `json:"data"`
}

// webhookData holds only the fields used to find the record.
type webhookData struct {
	Reference         string         `json:"reference"`
	MerchantReference string         `json:"merchantReference"`
	CustomerReference string         `json:"customerReference"`
	Metadata          map[string]any `json:"metadata"`
}

// outcome is a verified status ready to be recorded.
type outcome struct {
	entityID    string
	externalID  string
	status      commonv1.STATUS
	extras      data.JSONMap
	tenantID    string
	partitionID string
}

func (s *FincraWebhookServer) handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := util.Log(ctx).WithField("type", "fincra.webhook")
	defer logger.Release()

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxWebhookBody))
	if err != nil {
		s.metrics.WebhookRejected(ctx, "unknown", "decode_error")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	creds, err := s.resolveCredentials(ctx, r.URL.Query().Get("connection"))
	if err != nil || creds.WebhookSecret == "" || (creds.Connection != "" && creds.TenantID == "") {
		logger.WithError(err).Error("no webhook secret to verify Fincra webhook, rejecting")
		s.metrics.WebhookRejected(ctx, "unknown", "verification_failed")
		http.Error(w, "webhook verification unavailable", http.StatusServiceUnavailable)
		return
	}
	if !client.VerifyWebhookSignature(body, r.Header.Get(client.HeaderWebhookSignature), creds.WebhookSecret) {
		logger.Error("webhook signature verification failed")
		s.metrics.WebhookRejected(ctx, "unknown", "verification_failed")
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var event webhookEvent
	if err = json.Unmarshal(body, &event); err != nil {
		s.metrics.WebhookRejected(ctx, "unknown", "decode_error")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	kind := kindFromEvent(event.Event)
	if kind == "" {
		// Conversions and permanent virtual accounts are not ours to record.
		logger.WithField("event", event.Event).Debug("ignoring Fincra event")
		writeOK(w)
		return
	}
	s.metrics.WebhookReceived(ctx, kind)

	var res *outcome
	var code int
	if kind == kindPayout {
		res, code = s.verifyPayout(ctx, creds, &event)
	} else {
		res, code = s.verifyCollection(ctx, creds, &event)
	}
	if res == nil {
		http.Error(w, http.StatusText(code), code)
		return
	}
	tenantID, partitionID, ok := attribute(r, creds, res)
	if !ok {
		logger.WithField("connection", creds.Connection).Error("webhook tenancy differs from its connection's, rejecting")
		s.metrics.WebhookRejected(ctx, kind, "tenancy_mismatch")
		http.Error(w, "tenancy mismatch", http.StatusForbidden)
		return
	}
	res.tenantID, res.partitionID = tenantID, partitionID
	s.record(w, r, kind, res)
}

// resolveCredentials returns the named connection's credentials, or the
// service's own account when none is named. A named connection that cannot
// be resolved is an error, never a fallback to the service's account.
func (s *FincraWebhookServer) resolveCredentials(ctx context.Context, connection string) (*client.Credentials, error) {
	if connection != "" {
		return s.credsResolver.FromConnection(ctx, connection)
	}
	return s.credsResolver.Default()
}

// attribute decides the tenancy a verified outcome is recorded under.
//
// A webhook verified with a tenant's own connection was signed with a secret
// that tenant holds, so it can only speak for that connection's bound
// tenancy: any other tenancy in Fincra metadata or the URL is refused. Only
// the service's own account, whose secret no tenant holds, takes tenancy
// from the record's metadata or the webhook URL.
func attribute(r *http.Request, creds *client.Credentials, res *outcome) (string, string, bool) {
	tenantID, partitionID := res.tenantID, res.partitionID
	if tenantID == "" {
		tenantID = r.URL.Query().Get("tenant_id")
	}
	if partitionID == "" {
		partitionID = r.URL.Query().Get("partition_id")
	}
	if creds.Connection == "" {
		return tenantID, partitionID, true
	}
	if (tenantID != "" && tenantID != creds.TenantID) || (partitionID != "" && partitionID != creds.PartitionID) {
		return "", "", false
	}
	return creds.TenantID, creds.PartitionID, true
}

// kindFromEvent maps a Fincra event name to what it reports on.
func kindFromEvent(event string) string {
	lower := strings.ToLower(event)
	switch {
	case strings.HasPrefix(lower, "charge."), strings.HasPrefix(lower, "collection."):
		return kindCollection
	case strings.HasPrefix(lower, "payout."):
		return kindPayout
	}
	return ""
}

// collectionReferences are our candidate references for a collection
// event: the prompt id we stored in metadata, then the merchant reference,
// then the event reference unless Fincra generated it.
func collectionReferences(d *webhookData) []string {
	var refs []string
	add := func(ref string) {
		ref = strings.TrimSpace(ref)
		if ref == "" || strings.HasPrefix(strings.ToLower(ref), fincraReferencePrefix) {
			return
		}
		for _, have := range refs {
			if have == ref {
				return
			}
		}
		refs = append(refs, ref)
	}
	if v, ok := d.Metadata["prompt_id"].(string); ok {
		add(v)
	}
	add(d.MerchantReference)
	add(d.Reference)
	return refs
}

// verifyCollection reads the checkout or direct charge the event is about.
func (s *FincraWebhookServer) verifyCollection(
	ctx context.Context,
	creds *client.Credentials,
	event *webhookEvent,
) (*outcome, int) {
	logger := util.Log(ctx).WithField("type", "fincra.webhook.collection")
	defer logger.Release()

	refs := collectionReferences(&event.Data)
	if len(refs) == 0 {
		logger.Warn("collection event carries none of our references, ignoring")
		s.metrics.WebhookRejected(ctx, kindCollection, "unknown_payment")
		return nil, http.StatusNotFound
	}

	for _, ref := range refs {
		cp, err := s.cli.GetCheckout(ctx, creds, ref)
		if err == nil {
			return checkoutOutcome(ref, cp), http.StatusOK
		}
		if !client.IsNotFound(err) {
			logger.WithError(err).Error("could not verify checkout with Fincra")
			s.metrics.WebhookRejected(ctx, kindCollection, "verification_failed")
			return nil, http.StatusBadGateway
		}
		ch, err := s.cli.GetCharge(ctx, creds, ref)
		if err == nil {
			return chargeOutcome(ref, ch, creds), http.StatusOK
		}
		if !client.IsNotFound(err) {
			logger.WithError(err).Error("could not verify charge with Fincra")
			s.metrics.WebhookRejected(ctx, kindCollection, "verification_failed")
			return nil, http.StatusBadGateway
		}
	}
	logger.WithField("references", refs).Error("collection unknown to Fincra, rejecting")
	s.metrics.WebhookRejected(ctx, kindCollection, "unknown_payment")
	return nil, http.StatusNotFound
}

func checkoutOutcome(ref string, cp *client.CheckoutPayment) *outcome {
	status := client.CollectionStatus(cp.Status, cp.VarianceType)
	entityID := cp.MerchantReference
	if entityID == "" {
		entityID = ref
	}
	extras := data.JSONMap{
		"entity_type":             "prompt",
		collection.ExtraProvider:  providerName,
		collection.ExtraMode:      client.CollectionModeCheckout,
		"pay_code":                cp.Reference,
		"status":                  cp.Status,
		"provider_transaction_id": cp.Reference,
	}
	addVariance(extras, cp.VarianceType, cp.AmountExpected.Float(), cp.AmountReceived.Float())
	switch status {
	case commonv1.STATUS_SUCCESSFUL:
		if v := cp.Amount.Float(); v > 0 {
			extras["amount"] = formatAmount(v)
		}
		extras["currency"] = cp.Currency
	case commonv1.STATUS_FAILED:
		addFailure(extras, cp.Status, cp.VarianceType, cp.Message)
	default:
	}
	tenantID, _ := cp.Metadata["tenant_id"].(string)
	partitionID, _ := cp.Metadata["partition_id"].(string)
	return &outcome{
		entityID: entityID, externalID: cp.Reference, status: status, extras: extras,
		tenantID: tenantID, partitionID: partitionID,
	}
}

func chargeOutcome(ref string, ch *client.Charge, creds *client.Credentials) *outcome {
	status := client.CollectionStatus(ch.Status, ch.VarianceType)
	entityID := ch.Reference
	if entityID == "" {
		entityID = ref
	}
	extras := data.JSONMap{
		"entity_type":             "prompt",
		collection.ExtraProvider:  providerName,
		collection.ExtraMode:      client.CollectionModeDirect,
		collection.ExtraChargeID:  ch.ID.String(),
		"charge_type":             ch.Type,
		"status":                  ch.Status,
		"provider_transaction_id": firstNonEmpty(ch.ChargeReference, ch.ID.String()),
	}
	addVariance(extras, ch.VarianceType, ch.AmountExpected.Float(), ch.AmountReceived.Float())
	switch status {
	case commonv1.STATUS_SUCCESSFUL:
		if paid := ch.PaidAmount(creds.ResolveFeeBearer()); paid > 0 {
			extras["amount"] = formatAmount(paid)
			extras["currency"] = ch.Currency
		}
	case commonv1.STATUS_FAILED:
		addFailure(extras, ch.Status, ch.VarianceType, ch.Message)
	default:
	}
	return &outcome{entityID: entityID, externalID: ch.ID.String(), status: status, extras: extras}
}

// verifyPayout reads the payout the event is about.
func (s *FincraWebhookServer) verifyPayout(
	ctx context.Context,
	creds *client.Credentials,
	event *webhookEvent,
) (*outcome, int) {
	logger := util.Log(ctx).WithField("type", "fincra.webhook.payout")
	defer logger.Release()

	ref := strings.TrimSpace(event.Data.CustomerReference)
	if ref == "" {
		logger.Warn("payout event carries no customer reference, ignoring")
		s.metrics.WebhookRejected(ctx, kindPayout, "unknown_payment")
		return nil, http.StatusNotFound
	}
	p, err := s.cli.GetPayout(ctx, creds, ref)
	if client.IsNotFound(err) {
		logger.WithField("customer_reference", ref).Error("payout unknown to Fincra, rejecting")
		s.metrics.WebhookRejected(ctx, kindPayout, "unknown_payment")
		return nil, http.StatusNotFound
	}
	if err != nil {
		logger.WithError(err).Error("could not verify payout with Fincra")
		s.metrics.WebhookRejected(ctx, kindPayout, "verification_failed")
		return nil, http.StatusBadGateway
	}

	status := client.PayoutStatus(p.Status)
	extras := data.JSONMap{
		"entity_type":             "payment",
		collection.ExtraProvider:  providerName,
		"payout_id":               p.ID.String(),
		"status":                  p.Status,
		"provider_transaction_id": p.Reference,
		"customer_reference":      ref,
	}
	if p.AmountReceived > 0 {
		extras["amount_received"] = formatAmount(p.AmountReceived.Float())
	}
	if p.DestinationCurrency != "" {
		extras["currency"] = p.DestinationCurrency
	}
	if p.TraceID != "" {
		extras["trace_id"] = p.TraceID
	}
	if status == commonv1.STATUS_FAILED {
		reason := firstNonEmpty(p.Reason, p.Message, "The payout failed")
		extras["failure_code"] = "PAYOUT_FAILED"
		extras["failure_message"] = reason
	}
	return &outcome{entityID: ref, externalID: p.Reference, status: status, extras: extras}, http.StatusOK
}

// record stores a verified outcome. A record Fincra still reports in
// flight is answered 503, so Fincra delivers the webhook again later
// instead of the update being lost.
func (s *FincraWebhookServer) record(w http.ResponseWriter, r *http.Request, kind string, res *outcome) {
	logger := util.Log(r.Context()).
		WithField("type", "fincra.webhook."+kind).
		WithField("entity_id", res.entityID).
		WithField("status", res.status.String())
	defer logger.Release()

	if res.status == commonv1.STATUS_IN_PROCESS {
		logger.Warn("Fincra still reports the record in flight; asking for redelivery")
		s.metrics.WebhookRejected(r.Context(), kind, "not_final")
		http.Error(w, "record not final yet", http.StatusServiceUnavailable)
		return
	}

	ctx := tenantContext(r.Context(), res.tenantID, res.partitionID)
	_, err := s.paymentCli.StatusUpdate(ctx, &commonv1.StatusUpdateRequest{
		Id:         res.entityID,
		State:      client.StateFor(res.status),
		Status:     res.status,
		ExternalId: res.externalID,
		Extras:     res.extras.ToProtoStruct(),
	})
	if err != nil {
		logger.WithError(err).Error("could not update payment status")
		s.metrics.WebhookRejected(r.Context(), kind, "status_update_error")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeOK(w)
}

// tenantContext builds the claims the status update is recorded under.
func tenantContext(ctx context.Context, tenantID, partitionID string) context.Context {
	if tenantID == "" && partitionID == "" {
		return ctx
	}
	claims := &security.AuthenticationClaims{TenantID: tenantID, PartitionID: partitionID}
	return claims.ClaimsToContext(ctx)
}

func addVariance(extras data.JSONMap, varianceType string, expected, received float64) {
	if varianceType == "" {
		return
	}
	extras["variance_type"] = varianceType
	extras["amount_expected"] = formatAmount(expected)
	extras["amount_received"] = formatAmount(received)
}

func addFailure(extras data.JSONMap, status, varianceType, message string) {
	code := strings.ToUpper(status)
	msg := "The payment " + strings.ToLower(status)
	if strings.EqualFold(varianceType, client.VarianceUnderpayment) {
		code = client.FailureAmountMismatch
		msg = "Less than the amount due was paid; Fincra returns it to the payer"
	} else if m := strings.TrimSpace(message); m != "" && !strings.Contains(m, "<") {
		msg = m
	}
	extras["failure_code"] = code
	extras["failure_message"] = msg
}

func formatAmount(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func writeOK(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
