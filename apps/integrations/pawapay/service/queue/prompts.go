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

	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/credentials"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/antinvestor/service-payments/pkg/events"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/pitabwire/frame/v3/queue"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/proto"
)

type promptHandler struct {
	reporter
	credsResolver *credentials.Resolver
	pawapayCli    client.PawapayClient
}

// NewPromptHandler creates the queue worker for deposit (collection)
// prompts. A pawaPay deposit triggers a PIN prompt on the customer's phone,
// equivalent to an M-Pesa STK push. The deposit id is derived from the
// prompt id, so a redelivered prompt finds the deposit pawaPay already
// holds instead of prompting the customer twice.
func NewPromptHandler(
	eventsMan frameEvents.Manager,
	pawapayCli client.PawapayClient,
	credsResolver *credentials.Resolver,
) queue.SubscribeWorker {
	return &promptHandler{
		reporter: reporter{
			emitter:    events.NewStatusEmitter(eventsMan),
			metrics:    integrationobs.NewMetrics(providerName),
			entityType: events.EntityTypePrompt,
			kind:       "deposit",
		},
		credsResolver: credsResolver,
		pawapayCli:    pawapayCli,
	}
}

// lookup reports the deposit pawaPay holds, or found=false when it has none.
func (h *promptHandler) lookup(
	ctx context.Context,
	creds *client.Credentials,
	promptID, depositID string,
) (bool, error) {
	result, err := h.pawapayCli.GetDeposit(ctx, creds, depositID)
	if err != nil {
		return false, err
	}
	if result.Status != client.SearchStatusFound || result.Data == nil {
		return false, nil
	}
	d := result.Data
	provider := ""
	if d.Payer != nil {
		provider = d.Payer.AccountDetails.Provider
	}
	return true, h.record(ctx, promptID, d.DepositID, d.Status, provider, d.Created, d.FailureReason)
}

func (h *promptHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("type", "pawapay.prompt")
	defer logger.Release()

	prompt := paymentv1.InitiatePromptRequest{}
	if err := proto.Unmarshal(payload, &prompt); err != nil {
		logger.WithError(err).Error("failed to unmarshal prompt")
		h.metrics.QueueFailed(ctx, h.entityType, "unmarshal_error")
		return nil // non-retriable
	}

	promptID := prompt.GetId()
	depositID := paymentUUID(promptID)
	logger = logger.WithField("prompt_id", promptID).WithField("deposit_id", depositID)

	creds, err := h.credsResolver.FromHeaders(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve credentials")
		return h.fail(ctx, promptID, "", "credentials_error", map[string]any{"error": err.Error()})
	}

	// A redelivered prompt answers with the deposit already initiated.
	if found, lookErr := h.lookup(ctx, creds, promptID, depositID); lookErr != nil {
		logger.WithError(lookErr).Warn("deposit lookup failed, initiating")
	} else if found {
		logger.Debug("deposit already initiated")
		return nil
	}

	phoneNumber := prompt.GetRecipient().GetContactId()
	if phoneNumber == "" {
		phoneNumber = prompt.GetSource().GetContactId()
	}
	extraProvider := extraString(prompt.GetExtra(), "provider")
	provider, phoneNumber, err := resolveProvider(ctx, h.pawapayCli, creds, extraProvider, phoneNumber)
	if err != nil {
		if collection.IsUncertain(err) {
			return fmt.Errorf("deposit %s provider prediction: %w", promptID, err)
		}
		logger.WithError(err).Error("failed to resolve provider")
		return h.fail(ctx, promptID, "", "provider_error", map[string]any{"error": err.Error()})
	}

	depositReq := &client.DepositRequest{
		DepositID:            depositID,
		Amount:               formatMoneyAmount(prompt.GetAmount()),
		Currency:             prompt.GetAmount().GetCurrencyCode(),
		PhoneNumber:          phoneNumber,
		Provider:             provider,
		PreAuthorisationCode: extraString(prompt.GetExtra(), "pre_authorisation_code"),
		ClientReferenceID:    promptID,
		CustomerMessage:      sanitizeCustomerMessage(extraString(prompt.GetExtra(), "customer_message")),
		Metadata:             paymentMetadata(events.EntityTypePrompt, promptID, prompt.GetAmount(), headers),
	}

	resp, err := h.pawapayCli.InitiateDeposit(ctx, creds, depositReq)
	switch {
	case err != nil && !collection.IsUncertain(err):
		logger.WithError(err).Error("deposit refused by pawaPay")
		return h.fail(ctx, promptID, "", "provider_error", map[string]any{"error": err.Error()})
	case err != nil || resp.Status == client.InitiationStatusDuplicateIgnored:
		// Unknown outcome or already held: report what pawaPay holds, or
		// redeliver until it can be found or is refused.
		found, lookErr := h.lookup(ctx, creds, promptID, depositID)
		if lookErr != nil || !found {
			if err == nil {
				err = fmt.Errorf("deposit %s duplicate but not found", depositID)
			}
			logger.WithError(err).Warn("deposit outcome unknown; redelivering")
			return fmt.Errorf("deposit %s outcome unknown: %w", promptID, err)
		}
		return nil
	case resp.Status == client.InitiationStatusRejected:
		logger.WithField("failure_reason", resp.FailureReason).Error("deposit rejected")
		return h.fail(ctx, promptID, depositID, "rejected", map[string]any{
			"failure_code":    failureCode(resp.FailureReason),
			"failure_message": failureMessage(resp.FailureReason),
		})
	}

	logger.WithField("status", resp.Status).Debug("deposit initiated")
	return h.record(ctx, promptID, depositID, resp.Status, provider, resp.Created, nil)
}
