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
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/v2/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/config"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/coreapi"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/events"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/pitabwire/frame/v3/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	providerName = "jenga"

	defaultCurrency    = "KES"
	defaultCountryCode = "KE"
	defaultTelco       = "Safaricom"
	defaultPushType    = "STK"
	dateFormat         = "2006-01-02"

	// Failure codes for problems found before Jenga is called.
	failureCredentials    = "CREDENTIALS_ERROR"
	failureInvalidAmount  = "INVALID_AMOUNT"
	failureMissingPhone   = "MISSING_PHONE_NUMBER"
	failureMissingAccount = "MISSING_MERCHANT_ACCOUNT"
	failureProviderError  = "PROVIDER_ERROR"
)

// statusReader is the part of the payment service the worker uses to find a
// prompt's recorded status before pushing again.
type statusReader interface {
	Status(ctx context.Context, req *commonv1.StatusRequest) (*commonv1.StatusResponse, error)
}

type promptHandler struct {
	credentialResolver
	emitter      *events.StatusEmitter
	jengaCli     coreapi.JengaApiClient
	statusReader statusReader
	metrics      *integrationobs.Metrics
}

// NewPromptHandler creates the queue worker for STK/USSD push prompts.
//
// The prompt id is the key everywhere: the push is sent with the prompt's
// transaction reference, the callback URL carries the prompt id, and the
// reference plus requested amount are recorded on the prompt's status so
// the callback can be bound back to it.
func NewPromptHandler(
	eventsMan frameEvents.Manager,
	jengaCli coreapi.JengaApiClient,
	paymentCli paymentv1connect.PaymentServiceClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.JengaConfig,
) queue.SubscribeWorker {
	h := &promptHandler{
		credentialResolver: credentialResolver{settingsCli: settingsCli, cfg: cfg},
		emitter:            events.NewStatusEmitter(eventsMan),
		jengaCli:           jengaCli,
		metrics:            integrationobs.NewMetrics(providerName),
	}
	if paymentCli != nil {
		h.statusReader = paymentCli
	}
	return h
}

func (h *promptHandler) fail(ctx context.Context, promptID, reason, code, message string, extras map[string]any) error {
	h.metrics.QueueFailed(ctx, "prompt", reason)
	if extras == nil {
		extras = map[string]any{}
	}
	extras[collection.ExtraProvider] = providerName
	extras["failure_code"] = code
	extras["failure_message"] = message
	return h.emitter.Fail(ctx, events.EntityTypePrompt, promptID, message, extras)
}

func (h *promptHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("handler", "jenga.prompt")
	defer logger.Release()

	prompt := &paymentv1.InitiatePromptRequest{}
	if err := proto.Unmarshal(payload, prompt); err != nil {
		logger.WithError(err).Error("failed to unmarshal prompt")
		h.metrics.QueueFailed(ctx, "prompt", "unmarshal_error")
		return nil // non-retriable
	}
	promptID := prompt.GetId()
	logger = logger.WithField("prompt_id", promptID)

	// Redelivered message: a push already went out for this prompt. Pushing
	// again would re-bind the prompt and reject the first push's callback.
	prior := h.priorBinding(ctx, promptID)
	if prior.transactionID != "" {
		logger.WithField(models.ExtraTransactionID, prior.transactionID).Info("STK push already issued for prompt, skipping")
		h.metrics.QueueProcessed(ctx, "prompt")
		return nil
	}

	creds, err := h.extractCredentials(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve Jenga credentials")
		return h.fail(ctx, promptID, "credentials_error", failureCredentials, "The payment provider is not configured", nil)
	}

	if _, err = collection.MinorUnits(prompt.GetAmount()); err != nil {
		return h.fail(ctx, promptID, "rejected", failureInvalidAmount, "The amount must be positive with at most two decimals", nil)
	}
	amount := collection.DecimalString(prompt.GetAmount())
	currency := strings.ToUpper(strings.TrimSpace(prompt.GetAmount().GetCurrencyCode()))
	if currency == "" {
		currency = defaultCurrency
	}

	phone := strings.TrimSpace(prompt.GetSource().GetContactId())
	if phone == "" {
		phone = strings.TrimSpace(prompt.GetRecipient().GetContactId())
	}
	if phone == "" {
		return h.fail(ctx, promptID, "rejected", failureMissingPhone, "A mobile number is required for an STK push", nil)
	}

	extra := prompt.GetExtra()
	merchant := merchantAccount(prompt.GetRecipientAccount(), extra, creds)
	if merchant.AccountNumber == "" {
		return h.fail(ctx, promptID, "rejected", failureMissingAccount, "The merchant account to credit is missing", nil)
	}

	// The transaction reference and callback secret are unguessable and
	// never derived from the prompt id, which checkout exposes. A QUEUED
	// status from an earlier delivery already holds them; reuse those so the
	// callback of a push that did go out still binds.
	transactionRef, secret := prior.transactionRef, prior.secret
	if transactionRef == "" {
		transactionRef = extraString(extra, models.ExtraTransactionRef, collection.ExtraTxRef)
	}
	if transactionRef == "" || secret == "" {
		var err error
		if transactionRef == "" {
			if transactionRef, err = randomToken(transactionRefBytes); err != nil {
				return fmt.Errorf("mint transaction reference: %w", err)
			}
		}
		if secret, err = randomToken(callbackSecretBytes); err != nil {
			return fmt.Errorf("mint callback secret: %w", err)
		}
	}
	connection := headers[config.HeaderConnectionCredentials]
	binding := map[string]any{
		collection.ExtraProvider:      providerName,
		models.ExtraTransactionRef:    transactionRef,
		models.ExtraCallbackSecret:    secret,
		models.ExtraRequestedAmount:   amount,
		models.ExtraRequestedCurrency: currency,
		"mobile_number":               phone,
	}
	if connection != "" {
		binding[models.ExtraCredentialsConnection] = connection
	}

	// The binding is recorded before the push so a callback can be matched
	// even if the IN_PROCESS update after the push is lost.
	if err = h.emitter.Emit(ctx, events.EntityTypePrompt, promptID, "", commonv1.STATUS_QUEUED, cloneExtras(binding)); err != nil {
		return err
	}

	stkRequest := models.STKUSSDRequest{
		Merchant: merchant,
		Payment: models.Payment{
			Ref:          transactionRef,
			Amount:       amount,
			Currency:     currency,
			Telco:        firstNonEmpty(extraString(extra, "telco", collection.ExtraNetwork), defaultTelco),
			MobileNumber: phone,
			Date:         time.Now().Format(dateFormat),
			CallBackUrl:  stkCallbackURL(creds.CallbackURL, headers, promptID, secret),
			PushType:     firstNonEmpty(extraString(extra, "push_type", "pushType"), defaultPushType),
		},
		ID: promptID,
	}
	logger = logger.WithFields(map[string]any{
		models.ExtraTransactionRef: transactionRef,
		"amount":                   amount,
		"currency":                 currency,
	})

	apiCreds := toAPICreds(creds)
	token, err := h.jengaCli.GenerateBearerToken(ctx, apiCreds)
	if err != nil {
		if collection.IsUncertain(err) {
			logger.WithError(err).Warn("could not reach Jenga for a token; redelivering")
			return fmt.Errorf("jenga token for prompt %s: %w", promptID, err)
		}
		logger.WithError(err).Error("Jenga refused the credentials")
		return h.fail(ctx, promptID, "credentials_error", failureCredentials, "The payment provider refused the credentials", nil)
	}

	resp, err := h.jengaCli.InitiateSTKUSSD(ctx, apiCreds, stkRequest, token.AccessToken)
	if err != nil {
		if coreapi.IsRejected(err) {
			logger.WithError(err).Error("STK push refused by Jenga")
			return h.fail(ctx, promptID, "provider_error", failureProviderError, err.Error(), cloneExtras(binding))
		}
		// Outcome unknown (timeout, 5xx): Jenga may have pushed. Redeliver;
		// the callback still binds by transaction reference either way.
		logger.WithError(err).Warn("STK push outcome unknown; redelivering")
		return fmt.Errorf("stk push for prompt %s outcome unknown: %w", promptID, err)
	}

	logger.WithField(models.ExtraTransactionID, resp.TransactionID).Info("STK/USSD push initiated")

	inProcess := cloneExtras(binding)
	inProcess[models.ExtraTransactionID] = resp.TransactionID
	inProcess["response_code"] = resp.Code
	inProcess["message"] = resp.Message
	if err = h.emitter.Emit(ctx, events.EntityTypePrompt, promptID, resp.TransactionID, commonv1.STATUS_IN_PROCESS, inProcess); err != nil {
		return err
	}
	h.metrics.QueueProcessed(ctx, "prompt")
	return nil
}

const (
	transactionRefBytes = 12
	callbackSecretBytes = 16
)

// randomToken returns n random bytes as hex.
func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// priorBinding is what an earlier delivery recorded on the prompt.
type priorBinding struct {
	transactionID  string
	transactionRef string
	secret         string
}

// priorBinding reads the binding an earlier delivery recorded, if any. A
// lookup failure counts as "none" (the payment service reports a missing
// status as an error).
func (h *promptHandler) priorBinding(ctx context.Context, promptID string) priorBinding {
	if h.statusReader == nil || promptID == "" {
		return priorBinding{}
	}
	extras, _ := structpb.NewStruct(map[string]any{events.ExtraEntityType: events.EntityTypePrompt})
	resp, err := h.statusReader.Status(ctx, &commonv1.StatusRequest{Id: promptID, Extras: extras})
	if err != nil {
		util.Log(ctx).WithError(err).WithField("prompt_id", promptID).Debug("no prior prompt status found before STK push")
		return priorBinding{}
	}
	fields := resp.GetExtras().GetFields()
	return priorBinding{
		transactionID:  fields[models.ExtraTransactionID].GetStringValue(),
		transactionRef: fields[models.ExtraTransactionRef].GetStringValue(),
		secret:         fields[models.ExtraCallbackSecret].GetStringValue(),
	}
}

// merchantAccount is the account the STK push credits: the prompt's
// recipient account, else the prompt extras, else the merchant code.
func merchantAccount(account *paymentv1.Account, extra *structpb.Struct, creds *JengaCredentials) models.Merchant {
	m := models.Merchant{
		AccountNumber: firstNonEmpty(account.GetAccountNumber(), extraString(extra, "merchant_account_number", "account_number"), creds.MerchantCode),
		CountryCode:   firstNonEmpty(account.GetCountryCode(), extraString(extra, "country_code", collection.ExtraCountry), defaultCountryCode),
		Name:          firstNonEmpty(account.GetName(), extraString(extra, "merchant_name", "account_name")),
	}
	return m
}

// stkCallbackURL is this service's STK callback endpoint carrying the
// tenancy, credential connection, prompt id and per-prompt secret the
// callback must present to be bound.
func stkCallbackURL(base string, headers map[string]string, promptID, secret string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" {
		return ""
	}
	q := url.Values{}
	if v := headers["tenant_id"]; v != "" {
		q.Set("tenant_id", v)
	}
	if v := headers["partition_id"]; v != "" {
		q.Set("partition_id", v)
	}
	if v := headers[config.HeaderConnectionCredentials]; v != "" {
		q.Set("connection", v)
	}
	if promptID != "" {
		q.Set(models.CallbackParamPromptID, promptID)
	}
	if secret != "" {
		q.Set(models.CallbackParamSecret, secret)
	}
	u := base + models.STKCallbackPath
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	return u
}

// toAPICreds converts queue JengaCredentials to the coreapi.Credentials used by the API client.
func toAPICreds(creds *JengaCredentials) *coreapi.Credentials {
	return &coreapi.Credentials{
		MerchantCode:   creds.MerchantCode,
		ConsumerSecret: creds.ConsumerSecret,
		APIKey:         creds.APIKey,
		Environment:    creds.Environment,
		BaseURL:        creds.BaseURL,
		PrivateKeyPEM:  creds.PrivateKeyPEM,
		PrivateKeyPath: creds.PrivateKeyPath,
	}
}

// extraString returns the first non-empty string among the extras' keys.
func extraString(extra *structpb.Struct, keys ...string) string {
	if extra == nil {
		return ""
	}
	for _, key := range keys {
		if v, ok := extra.GetFields()[key]; ok {
			if s := strings.TrimSpace(v.GetStringValue()); s != "" {
				return s
			}
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func cloneExtras(in map[string]any) map[string]any {
	out := make(map[string]any, len(in)+4)
	for k, v := range in {
		out[k] = v
	}
	return out
}
