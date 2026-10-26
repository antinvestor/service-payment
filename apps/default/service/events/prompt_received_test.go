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
	"testing"

	"github.com/pitabwire/frame/v3/data"
	"github.com/pitabwire/util/decimalx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/antinvestor/service-payments/apps/default/service/models"
)

func successfulPrompt() *models.Prompt {
	amount := decimalx.NewFromInt64(30000)
	p := &models.Prompt{
		SourceID:          "profile-payer",
		SourceProfileType: "Person",
		SourceContactID:   "contact-payer",
		RecipientID:       "profile-lender",
		Amount:            &amount,
		Extra: data.JSONMap{
			"currency":        "UGX",
			"transaction_ref": "ref-1",
			"product_id":      "prod-1",
			"loan_account_id": "loan-1",
		},
	}
	p.BaseModel.ID = "prompt-1"
	p.ID = "prompt-1"
	p.TenantID = "tenant-1"
	p.PartitionID = "partition-1"
	return p
}

func TestPaymentFromPromptIsTheMoneyThePromptCollected(t *testing.T) {
	status := &models.Status{Extra: data.JSONMap{"financial_transaction_id": "MTN-9"}}

	payment := paymentFromPrompt(successfulPrompt(), status)

	assert.Equal(t, "prompt-1", payment.GetID(), "a replayed success converges on one payment")
	assert.False(t, payment.OutBound)
	assert.Equal(t, "profile-payer", payment.SenderProfileID)
	assert.Equal(t, "contact-payer", payment.SenderContactID)
	assert.Equal(t, "profile-lender", payment.RecipientProfileID)
	require.NotNil(t, payment.Amount)
	assert.Equal(t, "30000", payment.Amount.String())
	assert.Equal(t, "UGX", payment.Currency)
	assert.Equal(t, "MTN-9", payment.TransactionID)
	assert.Equal(t, "tenant-1", payment.TenantID)
	assert.Equal(t, "partition-1", payment.PartitionID)
	assert.Equal(t, "prod-1", payment.Extra.GetString("product_id"))
	assert.Equal(t, "loan-1", payment.Extra.GetString("loan_account_id"))
	assert.Equal(t, "prompt-1", payment.Extra.GetString(extraPromptID))
}

func TestPaymentFromPromptFallsBackToThePromptReference(t *testing.T) {
	payment := paymentFromPrompt(successfulPrompt(), &models.Status{})
	assert.Equal(t, "ref-1", payment.TransactionID)
}
