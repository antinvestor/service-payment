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

type promptHandler struct {
	statusEmitter
	credsResolver   *credentials.Resolver
	cli             client.FincraClient
	cfg             *config.FincraConfig
	chargeOperators map[string]string
	metrics         *integrationobs.Metrics
}

// NewPromptHandler creates the queue worker for collection prompts.
//
// In checkout mode (the default) it creates a Fincra hosted checkout and
// hands the payer its link. In direct mode it starts a direct charge: mobile
// money prompts the payer's phone (asking checkout for an OTP when the
// operator needs one) and bank transfer returns an account to pay into.
// Prompt ids are Fincra references, so a redelivered prompt finds the
// existing checkout or charge instead of creating another.
func NewPromptHandler(
	eventsMan frameEvents.Manager,
	cli client.FincraClient,
	credsResolver *credentials.Resolver,
	cfg *config.FincraConfig,
) queue.SubscribeWorker {
	return &promptHandler{
		statusEmitter:   statusEmitter{eventsMan: eventsMan},
		credsResolver:   credsResolver,
		cli:             cli,
		cfg:             cfg,
		chargeOperators: parseCodeMap(cfg.ChargeOperators),
		metrics:         integrationobs.NewMetrics(providerName),
	}
}

// promptJob is one prompt being collected.
type promptJob struct {
	id          string
	creds       *client.Credentials
	extra       *structpb.Struct
	amount      int64
	currency    string
	phone       string
	customer    client.Customer
	tenantID    string
	partitionID string
	connection  string
}

func (h *promptHandler) fail(ctx context.Context, promptID, externalID, reason string, extras map[string]any) error {
	h.metrics.QueueFailed(ctx, "prompt", reason)
	extras["entity_type"] = "prompt"
	extras[collection.ExtraProvider] = providerName
	return h.emitStatus(ctx, promptID, externalID, commonv1.STATUS_FAILED, extras)
}

func (h *promptHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "fincra.prompt")
	defer logger.Release()

	prompt := &paymentv1.InitiatePromptRequest{}
	if err := proto.Unmarshal(payload, prompt); err != nil {
		logger.WithError(err).Error("failed to unmarshal prompt")
		h.metrics.QueueFailed(ctx, "prompt", "unmarshal_error")
		return nil // non-retriable
	}
	promptID := prompt.GetId()
	logger = logger.WithField("prompt_id", promptID)

	creds, err := h.credsResolver.FromHeaders(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		return h.fail(ctx, promptID, "", "credentials_error", failureExtras(failureCredentials, ""))
	}

	extra := prompt.GetExtra()
	if strings.EqualFold(extraString(extra, collection.ExtraAction), collection.ActionAuthorize) {
		return h.authorize(ctx, promptID, creds, extra)
	}

	amount, ok := wholeUnits(prompt.GetAmount())
	currency := strings.ToUpper(strings.TrimSpace(prompt.GetAmount().GetCurrencyCode()))
	if !ok || currency == "" {
		return h.fail(ctx, promptID, "", "invalid_amount", failureExtras(failureInvalidAmount, ""))
	}

	phone := normalizeMSISDN(prompt.GetSource().GetContactId())
	if phone == "" {
		phone = normalizeMSISDN(prompt.GetRecipient().GetContactId())
	}
	customer, err := buildCustomer(extra, phone, h.cfg.FallbackCustomerEmail)
	if err != nil {
		return h.fail(ctx, promptID, "", "invalid_customer", failureExtras(failureMissingEmail, ""))
	}

	tenantID, partitionID := tenancy(ctx, headers)
	job := &promptJob{
		id: promptID, creds: creds, extra: extra, amount: amount, currency: currency,
		phone: phone, customer: customer, tenantID: tenantID, partitionID: partitionID,
		connection: headers[config.HeaderConnectionCredentials],
	}

	method := collectionMethod(extra)
	if creds.ResolveCollectionMode() == client.CollectionModeDirect && method != client.ChargeTypeCard {
		return h.directCharge(ctx, job, method)
	}
	return h.hostedCheckout(ctx, job, method)
}

// --- hosted checkout ---

func (h *promptHandler) hostedCheckout(ctx context.Context, job *promptJob, method string) error {
	logger := util.Log(ctx).WithField("type", "fincra.prompt.checkout").WithField("prompt_id", job.id)
	defer logger.Release()

	// A redelivered prompt answers with the checkout it already created.
	if existing, err := h.cli.GetCheckout(ctx, job.creds, job.id); err == nil {
		logger.Debug("checkout already exists for prompt")
		return h.emitCheckout(ctx, job, existing.Reference, existing)
	} else if !client.IsNotFound(err) {
		logger.WithError(err).Warn("checkout lookup failed, creating")
	}

	req := &client.CheckoutRequest{
		Amount:                job.amount,
		Currency:              job.currency,
		Customer:              job.customer,
		Reference:             job.id,
		FeeBearer:             job.creds.ResolveFeeBearer(),
		RedirectURL:           h.redirectURL(job.extra),
		PaymentMethods:        h.checkoutMethods(method),
		SettlementDestination: "wallet",
		Metadata:              promptMetadata(job),
	}
	if len(req.PaymentMethods) == 1 {
		req.DefaultPaymentMethod = req.PaymentMethods[0]
	}

	link, err := h.cli.CreateCheckout(ctx, job.creds, req)
	if err != nil {
		if client.IsRejected(err) && !client.IsDuplicate(err) {
			logger.WithError(err).Error("checkout refused by Fincra")
			return h.fail(ctx, job.id, "", "provider_error", providerFailureExtras(err))
		}
		// Duplicate or unknown outcome: report what Fincra holds, or have
		// the prompt redelivered while it holds nothing yet.
		existing, getErr := h.cli.GetCheckout(ctx, job.creds, job.id)
		if getErr != nil {
			logger.WithError(err).Warn("checkout outcome unknown; redelivering")
			return fmt.Errorf("checkout %s outcome unknown: %w", job.id, err)
		}
		return h.emitCheckout(ctx, job, existing.Reference, existing)
	}

	payCode := link.PayCode
	url := link.Link
	if url == "" {
		url = job.creds.CheckoutLink(payCode)
	}
	h.metrics.QueueProcessed(ctx, "prompt")
	return h.emitStatus(ctx, job.id, payCode, commonv1.STATUS_IN_PROCESS, checkoutExtras(job, payCode, url))
}

// emitCheckout reports an existing checkout's state.
func (h *promptHandler) emitCheckout(ctx context.Context, job *promptJob, payCode string, cp *client.CheckoutPayment) error {
	status := client.CollectionStatus(cp.Status, cp.VarianceType)
	extras := checkoutExtras(job, payCode, job.creds.CheckoutLink(payCode))
	switch status {
	case commonv1.STATUS_FAILED:
		code := strings.ToUpper(cp.Status)
		if strings.EqualFold(cp.VarianceType, client.VarianceUnderpayment) {
			code = client.FailureAmountMismatch
		}
		for k, v := range failureExtras(code, "") {
			extras[k] = v
		}
		return h.fail(ctx, job.id, payCode, "rejected", extras)
	case commonv1.STATUS_SUCCESSFUL:
		delete(extras, collection.ExtraCheckoutURL)
		delete(extras, collection.ExtraNextAction)
		extras["amount"] = formatAmount(cp.Amount.Float())
		extras["currency"] = cp.Currency
	default:
	}
	h.metrics.QueueProcessed(ctx, "prompt")
	return h.emitStatus(ctx, job.id, payCode, status, extras)
}

func checkoutExtras(job *promptJob, payCode, url string) map[string]any {
	extras := map[string]any{
		"entity_type":            "prompt",
		collection.ExtraProvider: providerName,
		collection.ExtraMode:     client.CollectionModeCheckout,
		"pay_code":               payCode,
		"reference":              job.id,
		"requested_amount":       fmt.Sprintf("%d", job.amount),
		"requested_currency":     job.currency,
	}
	if url != "" {
		extras[collection.ExtraCheckoutURL] = url
		extras[collection.ExtraNextAction] = collection.NextActionRedirectURL
	}
	return extras
}

// checkoutMethods narrows the hosted checkout to the requested rail, or to
// the configured list.
func (h *promptHandler) checkoutMethods(method string) []string {
	if method != "" {
		return []string{method}
	}
	var out []string
	for _, m := range strings.Split(h.cfg.CheckoutPaymentMethods, ",") {
		if m = strings.TrimSpace(m); m != "" {
			out = append(out, m)
		}
	}
	return out
}

func (h *promptHandler) redirectURL(extra *structpb.Struct) string {
	if u := extraString(extra, collection.ExtraSuccessURL, collection.ExtraRedirectURL); u != "" {
		return u
	}
	return h.cfg.DefaultRedirectURL
}

// promptMetadata travels with the checkout and comes back on verify, so a
// webhook can be attributed to its prompt and tenancy from Fincra's record.
func promptMetadata(job *promptJob) map[string]string {
	meta := map[string]string{"prompt_id": job.id}
	if job.tenantID != "" {
		meta["tenant_id"] = job.tenantID
	}
	if job.partitionID != "" {
		meta["partition_id"] = job.partitionID
	}
	if ref := extraString(job.extra, collection.ExtraSessionRef); ref != "" {
		meta["session_ref"] = ref
	}
	return meta
}

// --- direct charge ---

func (h *promptHandler) directCharge(ctx context.Context, job *promptJob, method string) error {
	logger := util.Log(ctx).WithField("type", "fincra.prompt.charge").WithField("prompt_id", job.id)
	defer logger.Release()

	if method == "" {
		method = client.ChargeTypeBankTransfer
		if job.phone != "" {
			method = client.ChargeTypeMobileMoney
		}
	}

	// A redelivered prompt answers with the charge it already started.
	if existing, err := h.cli.GetCharge(ctx, job.creds, job.id); err == nil {
		logger.Debug("charge already exists for prompt")
		return h.emitCharge(ctx, job.id, job.creds, existing)
	} else if !client.IsNotFound(err) {
		logger.WithError(err).Warn("charge lookup failed, initiating")
	}

	req := &client.ChargeRequest{
		Type:                  method,
		Amount:                job.amount,
		Reference:             job.id,
		Currency:              job.currency,
		Customer:              job.customer,
		SettlementDestination: "wallet",
		CallbackURL:           h.redirectURL(job.extra),
		WebhookURL:            webhookURL(h.cfg.WebhookBaseURL, job.tenantID, job.partitionID, job.connection),
		Description:           extraString(job.extra, "description", collection.ExtraSessionRef),
	}

	switch method {
	case client.ChargeTypeMobileMoney:
		if job.phone == "" {
			return h.fail(ctx, job.id, "", "invalid_customer", failureExtras(failureMissingPhone, ""))
		}
		operator := h.chargeOperator(job.extra, job.currency, job.phone)
		if operator == "" {
			return h.fail(ctx, job.id, "", "invalid_operator", failureExtras(failureUnknownOperator, ""))
		}
		req.Phone = "+" + job.phone
		req.Operator = operator
	case client.ChargeTypeBankTransfer:
	default:
		return h.fail(ctx, job.id, "", "invalid_method", failureExtras(failureUnsupportedMethod, ""))
	}

	charge, err := h.cli.InitiateCharge(ctx, job.creds, req)
	if err != nil {
		if client.IsRejected(err) && !client.IsDuplicate(err) {
			logger.WithError(err).Error("charge refused by Fincra")
			return h.fail(ctx, job.id, "", "provider_error", providerFailureExtras(err))
		}
		existing, getErr := h.cli.GetCharge(ctx, job.creds, job.id)
		if getErr != nil {
			logger.WithError(err).Warn("charge outcome unknown; redelivering")
			return fmt.Errorf("charge %s outcome unknown: %w", job.id, err)
		}
		charge = existing
	}
	return h.emitCharge(ctx, job.id, job.creds, charge)
}

// chargeOperator picks the operator code for a mobile money charge: the
// payer's chosen network, the configured code for that network or the
// currency, then the network the number belongs to.
func (h *promptHandler) chargeOperator(extra *structpb.Struct, currency, phone string) string {
	network := strings.ToUpper(extraString(extra, collection.ExtraNetwork))
	if network == "" {
		network = detectNetwork(phone)
	}
	if code, ok := h.chargeOperators[network]; ok && network != "" {
		return code
	}
	if code, ok := h.chargeOperators[currency]; ok {
		return code
	}
	return network
}

// authorize completes a pending charge with the OTP the payer entered on
// the checkout page (portable action=authorize continuation).
func (h *promptHandler) authorize(ctx context.Context, promptID string, creds *client.Credentials, extra *structpb.Struct) error {
	logger := util.Log(ctx).WithField("type", "fincra.prompt.authorize").WithField("prompt_id", promptID)
	defer logger.Release()

	chargeID := extraString(extra, collection.ExtraChargeID)
	otp := extraString(extra, collection.ExtraOTP)
	if chargeID == "" || otp == "" {
		return h.fail(ctx, promptID, chargeID, "invalid_authorization", failureExtras(failureMissingCharge, ""))
	}
	charge, err := h.cli.AuthorizeCharge(ctx, creds, chargeID, &client.AuthorizeRequest{OTP: otp})
	if err != nil {
		logger.WithError(err).Error("charge authorization failed")
		return h.fail(ctx, promptID, chargeID, "provider_error", providerFailureExtras(err))
	}
	return h.emitCharge(ctx, promptID, creds, charge)
}

// emitCharge reports a direct charge's state with the details the checkout
// page needs to guide the payer.
func (h *promptHandler) emitCharge(ctx context.Context, promptID string, creds *client.Credentials, ch *client.Charge) error {
	status := client.CollectionStatus(ch.Status, ch.VarianceType)
	extras := chargeExtras(ch)
	if status == commonv1.STATUS_FAILED {
		code := strings.ToUpper(ch.Status)
		if strings.EqualFold(ch.VarianceType, client.VarianceUnderpayment) {
			code = client.FailureAmountMismatch
		}
		for k, v := range failureExtras(code, plainText(ch.Message)) {
			extras[k] = v
		}
		return h.fail(ctx, promptID, ch.ID.String(), "rejected", extras)
	}
	if status == commonv1.STATUS_SUCCESSFUL {
		addPaidAmount(extras, ch, creds)
	}
	h.metrics.QueueProcessed(ctx, "prompt")
	return h.emitStatus(ctx, promptID, ch.ID.String(), status, extras)
}

// addPaidAmount reports a successful charge's amount for checkout's paid
// amount check.
func addPaidAmount(extras map[string]any, ch *client.Charge, creds *client.Credentials) {
	if paid := ch.PaidAmount(creds.ResolveFeeBearer()); paid > 0 {
		extras["amount"] = formatAmount(paid)
		extras["currency"] = ch.Currency
	}
}

func chargeExtras(ch *client.Charge) map[string]any {
	extras := map[string]any{
		"entity_type":            "prompt",
		collection.ExtraProvider: providerName,
		collection.ExtraMode:     client.CollectionModeDirect,
		collection.ExtraChargeID: ch.ID.String(),
		"charge_type":            ch.Type,
		"status":                 ch.Status,
		"reference":              ch.Reference,
	}
	if ch.ChargeReference != "" {
		extras["charge_reference"] = ch.ChargeReference
	}
	if ch.Fee > 0 {
		extras["fee"] = formatAmount(ch.Fee.Float())
	}
	if msg := plainText(ch.Message); msg != "" {
		extras[collection.ExtraPaymentInstruction] = msg
	}
	if strings.EqualFold(ch.AuthModel, client.AuthModelOTP) || strings.EqualFold(ch.Authorization.Mode, client.AuthModelOTP) {
		if isPending(ch.Status) {
			extras[collection.ExtraNextAction] = collection.NextActionRequiresOTP
			extras[collection.ExtraNextActionType] = collection.NextActionRequiresOTP
			extras[collection.ExtraAuthorizationType] = "otp"
		}
	}
	if va := ch.VirtualAccount; va != nil && va.AccountNumber != "" {
		extras[collection.ExtraBankName] = va.BankName
		extras[collection.ExtraBankAccountNumber] = va.AccountNumber
		if va.AccountName != "" {
			extras[collection.ExtraBankAccountName] = va.AccountName
		}
		if va.ExpiresAt != "" {
			extras[collection.ExtraPaymentExpiresAt] = va.ExpiresAt
		} else {
			// A pool account (no expiry) matches the payment by the charge id
			// in the transfer narration.
			extras[collection.ExtraPaymentReference] = ch.ID.String()
		}
		extras[collection.ExtraNextAction] = collection.NextActionBankTransfer
		extras[collection.ExtraPaymentInstruction] = bankInstruction(ch, va)
	}
	return extras
}

func bankInstruction(ch *client.Charge, va *client.VirtualAccount) string {
	amount := ch.AmountExpected.Float()
	if amount <= 0 {
		amount = ch.Amount.Float()
	}
	msg := fmt.Sprintf("Transfer exactly %s %s to %s account %s", ch.Currency, formatAmount(amount), va.BankName, va.AccountNumber)
	if va.AccountName != "" {
		msg += " (" + va.AccountName + ")"
	}
	if va.ExpiresAt == "" {
		msg += ", with " + ch.ID.String() + " as the transfer narration"
	} else {
		msg += " before " + va.ExpiresAt
	}
	return msg
}

func isPending(status string) bool {
	switch strings.ToLower(status) {
	case client.StatusPending, client.StatusProcessing, client.StatusInitiated, "":
		return true
	}
	return false
}
