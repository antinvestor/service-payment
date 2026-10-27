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

package events

import (
	"context"
	"fmt"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/pitabwire/frame/v3/data"
	"github.com/pitabwire/util"

	"github.com/antinvestor/service-payments/apps/default/service/models"
)

const (
	entityTypePrompt = "prompt"
	// extraPromptID names, on a payment received from a prompt, the prompt
	// that asked for it.
	extraPromptID = "prompt_id"
)

// receivePrompt records the money a successful prompt collected as an
// inbound payment, so it is routed like any other payment received: the
// service that asked the payer to pay learns the money arrived. The payment
// takes the prompt's id, so a replayed success status records one payment.
func (e *StatusSave) receivePrompt(ctx context.Context, status *models.Status) error {
	if status.EntityType != entityTypePrompt ||
		status.Status != int32(commonv1.STATUS_SUCCESSFUL.Number()) ||
		e.promptRepo == nil || e.eventMan == nil {
		return nil
	}

	prompt, err := e.promptRepo.GetByID(ctx, status.EntityID)
	if err != nil {
		return fmt.Errorf("load successful prompt %s: %w", status.EntityID, err)
	}
	if prompt.Amount == nil {
		util.Log(ctx).WithField("prompt_id", status.EntityID).
			Warn("successful prompt has no amount; no payment received")
		return nil
	}

	payment := paymentFromPrompt(prompt, status)
	if err = e.eventMan.Emit(ctx, EventNamePaymentSave, payment); err != nil {
		return fmt.Errorf("receive payment for prompt %s: %w", status.EntityID, err)
	}
	return nil
}

// paymentFromPrompt is the inbound payment a successful prompt collected:
// from the prompt's source to its recipient, for the amount asked, carrying
// the prompt's extras so the receiving service can tell what it was for.
func paymentFromPrompt(prompt *models.Prompt, status *models.Status) *models.Payment {
	extra := data.JSONMap{}
	for k, v := range prompt.Extra {
		extra[k] = v
	}
	id := promptID(prompt)
	extra[extraPromptID] = id

	payment := &models.Payment{
		SenderProfileType:    prompt.SourceProfileType,
		SenderProfileID:      prompt.SourceID,
		SenderContactID:      prompt.SourceContactID,
		RecipientProfileType: prompt.RecipientProfileType,
		RecipientProfileID:   prompt.RecipientID,
		RecipientContactID:   prompt.RecipientContactID,
		Amount:               prompt.Amount,
		Currency:             prompt.Extra.GetString("currency"),
		TransactionID:        promptTransactionID(prompt, status),
		ReferenceID:          prompt.Extra.GetString("reference_id"),
		OutBound:             false,
		Extra:                extra,
	}
	payment.CopyPartitionInfo(&prompt.BaseModel)
	payment.ID = id
	return payment
}

// promptID reads the prompt's id: the model declares its own ID column
// beside the embedded BaseModel's, and either may be the one loaded.
func promptID(prompt *models.Prompt) string {
	if prompt.ID != "" {
		return prompt.ID
	}
	return prompt.GetID()
}

// promptTransactionID is the provider's id for the money movement, or the
// prompt's own transaction reference when the provider gave none.
func promptTransactionID(prompt *models.Prompt, status *models.Status) string {
	for _, key := range []string{"financial_transaction_id", "transaction_id", "external_id"} {
		if v := status.Extra.GetString(key); v != "" {
			return v
		}
	}
	return prompt.Extra.GetString("transaction_ref")
}
