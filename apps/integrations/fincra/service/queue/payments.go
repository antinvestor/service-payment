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

package queue

import (
	"context"
	"fmt"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/v2/profile/v1/profilev1connect"
	profilev1 "buf.build/gen/go/antinvestor/profile/protocolbuffers/go/profile/v1"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/config"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/credentials"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/pitabwire/frame/v3/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type paymentHandler struct {
	statusEmitter
	credsResolver   *credentials.Resolver
	cli             client.FincraClient
	catalog         *client.Catalog
	profileCli      profilev1connect.ProfileServiceClient
	cfg             *config.FincraConfig
	payoutProviders map[string]string
	metrics         *integrationobs.Metrics
}

// NewPaymentHandler creates the queue worker for payouts. The payment
// service publishes a payout with the payee's verified mobile number as the
// recipient contact; Fincra pays that number's mobile money wallet. The
// payment id is the Fincra customerReference, so a redelivered payout finds
// the one already sent instead of paying twice.
func NewPaymentHandler(
	eventsMan frameEvents.Manager,
	cli client.FincraClient,
	catalog *client.Catalog,
	profileCli profilev1connect.ProfileServiceClient,
	credsResolver *credentials.Resolver,
	cfg *config.FincraConfig,
) queue.SubscribeWorker {
	return &paymentHandler{
		statusEmitter:   statusEmitter{eventsMan: eventsMan},
		credsResolver:   credsResolver,
		cli:             cli,
		catalog:         catalog,
		profileCli:      profileCli,
		cfg:             cfg,
		payoutProviders: parseCodeMap(cfg.PayoutProviders),
		metrics:         integrationobs.NewMetrics(providerName),
	}
}

func (h *paymentHandler) fail(ctx context.Context, paymentID, externalID, reason string, extras map[string]any) error {
	h.metrics.QueueFailed(ctx, "payment", reason)
	extras["entity_type"] = "payment"
	extras[collection.ExtraProvider] = providerName
	return h.emitStatus(ctx, paymentID, externalID, commonv1.STATUS_FAILED, extras)
}

func (h *paymentHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "fincra.payment")
	defer logger.Release()

	payment := &paymentv1.Payment{}
	if err := proto.Unmarshal(payload, payment); err != nil {
		logger.WithError(err).Error("failed to unmarshal payment")
		h.metrics.QueueFailed(ctx, "payment", "unmarshal_error")
		return nil // non-retriable
	}
	paymentID := payment.GetId()
	logger = logger.WithField("payment_id", paymentID)

	creds, err := h.credsResolver.FromHeaders(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		return h.fail(ctx, paymentID, "", "credentials_error", failureExtras(failureCredentials, ""))
	}

	// A redelivered payout answers with the one already sent.
	if existing, getErr := h.cli.GetPayout(ctx, creds, paymentID); getErr == nil {
		logger.Debug("payout already sent")
		return h.emitPayout(ctx, paymentID, existing)
	} else if !client.IsNotFound(getErr) {
		logger.WithError(getErr).Warn("payout lookup failed, sending")
	}

	phone := normalizeMSISDN(payment.GetRecipient().GetContactId())
	if phone == "" {
		return h.fail(ctx, paymentID, "", "invalid_recipient", failureExtras(failureMissingPhone, ""))
	}
	currency := strings.ToUpper(strings.TrimSpace(payment.GetAmount().GetCurrencyCode()))
	country := creds.Country
	if c, ok := corridorForPhone(phone); ok {
		country = c.Country
	}
	if currency == "" || country == "" {
		return h.fail(ctx, paymentID, "", "invalid_recipient", failureExtras(failureUnknownCountry, ""))
	}

	providerCode, err := h.payoutProvider(ctx, creds, country, currency, phone)
	if err != nil {
		logger.WithError(err).Error("could not load payout providers")
		return h.fail(ctx, paymentID, "", "provider_error", providerFailureExtras(err))
	}
	if providerCode == "" {
		return h.fail(ctx, paymentID, "", "invalid_operator", failureExtras(failureUnknownOperator, ""))
	}

	amount, ok := payoutAmount(payment.GetAmount(), requiresWholePayout(currency, providerCode))
	if !ok {
		return h.fail(ctx, paymentID, "", "invalid_amount", failureExtras(failureInvalidAmount,
			"The amount must be positive, with at most two decimal places (whole units for UGX and Airtel)"))
	}

	first, last, full := h.payeeName(ctx, payment.GetRecipient().GetProfileId())
	req := &client.PayoutRequest{
		Business:            creds.BusinessID,
		SourceCurrency:      currency,
		DestinationCurrency: currency,
		Amount:              amount,
		Description:         payoutDescription(payment),
		PaymentDestination:  client.PaymentDestinationMobileMoney,
		CustomerReference:   paymentID,
		Beneficiary: client.Beneficiary{
			FirstName:         first,
			LastName:          last,
			AccountHolderName: full,
			Type:              client.BeneficiaryIndividual,
			Country:           country,
			AccountNumber:     phone,
			MobileMoneyCode:   providerCode,
		},
	}

	payout, err := h.cli.CreatePayout(ctx, creds, req)
	if err != nil {
		if client.IsRejected(err) && !client.IsDuplicate(err) {
			logger.WithError(err).Error("payout refused by Fincra")
			return h.fail(ctx, paymentID, "", "provider_error", providerFailureExtras(err))
		}
		// Duplicate or unknown outcome (timeout, 5xx): Fincra may hold the
		// payout. Never report it failed on a guess — report what Fincra
		// holds, or redeliver until it can be found or is refused.
		existing, getErr := h.cli.GetPayout(ctx, creds, paymentID)
		if getErr != nil {
			logger.WithError(err).Warn("payout outcome unknown; redelivering")
			return fmt.Errorf("payout %s outcome unknown: %w", paymentID, err)
		}
		payout = existing
	}
	if payout.CustomerReference == "" {
		payout.CustomerReference = paymentID
	}
	return h.emitPayout(ctx, paymentID, payout)
}

// payoutProvider picks the mobileMoneyCode: configured for the payee's
// network or the currency, else the provider in Fincra's list that matches
// the network the number belongs to.
func (h *paymentHandler) payoutProvider(
	ctx context.Context,
	creds *client.Credentials,
	country, currency, phone string,
) (string, error) {
	network := detectNetwork(phone)
	if code, ok := h.payoutProviders[network]; ok && network != "" {
		return code, nil
	}
	if code, ok := h.payoutProviders[currency]; ok {
		return code, nil
	}
	providers, err := h.catalog.Providers(ctx, creds, country, currency, client.PaymentDestinationMobileMoney)
	if err != nil {
		return "", err
	}
	if p, ok := client.MatchProvider(providers, network); ok {
		return p.Code, nil
	}
	return "", nil
}

// payeeName reads the payee's name from their profile, falling back to the
// configured default. Fincra needs first and last names for individuals.
func (h *paymentHandler) payeeName(ctx context.Context, profileID string) (string, string, string) {
	full := ""
	if h.profileCli != nil && profileID != "" {
		resp, err := h.profileCli.GetById(ctx, &profilev1.GetByIdRequest{Id: profileID})
		if err != nil {
			util.Log(ctx).WithError(err).WithField("profile_id", profileID).Warn("could not read payee profile name")
		} else {
			full = profileName(resp.GetData().GetProperties())
		}
	}
	if full == "" {
		full = h.cfg.DefaultBeneficiaryName
	}
	first, last := splitName(full)
	return first, last, full
}

// profileName is the profile's display name (auth identity first).
func profileName(props *structpb.Struct) string {
	for _, key := range []string{"au_name", "name", "display_name", "displayName", "full_name"} {
		if n := extraString(props, key); n != "" {
			return n
		}
	}
	return strings.TrimSpace(extraString(props, "given_name") + " " + extraString(props, "family_name"))
}

// splitName splits a full name into first and last; a single word is used
// for both, since Fincra requires each.
func splitName(full string) (string, string) {
	parts := strings.Fields(full)
	switch len(parts) {
	case 0:
		return "", ""
	case 1:
		return parts[0], parts[0]
	}
	return strings.Join(parts[:len(parts)-1], " "), parts[len(parts)-1]
}

func payoutDescription(payment *paymentv1.Payment) string {
	if ref := strings.TrimSpace(payment.GetReferenceId()); ref != "" {
		return "Payout " + ref
	}
	return "Payout " + payment.GetId()
}

// emitPayout reports a payout's state.
func (h *paymentHandler) emitPayout(ctx context.Context, paymentID string, p *client.Payout) error {
	status := client.PayoutStatus(p.Status)
	extras := map[string]any{
		"entity_type":             "payment",
		collection.ExtraProvider:  providerName,
		"payout_id":               p.ID.String(),
		"status":                  p.Status,
		"provider_transaction_id": p.Reference,
		"customer_reference":      p.CustomerReference,
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
		reason := firstNonEmpty(p.Reason, p.Message)
		for k, v := range failureExtras("PAYOUT_FAILED", reason) {
			extras[k] = v
		}
		return h.fail(ctx, paymentID, p.Reference, "rejected", extras)
	}
	h.metrics.QueueProcessed(ctx, "payment")
	return h.emitStatus(ctx, paymentID, p.Reference, status, extras)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
