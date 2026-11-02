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

package client

import (
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
)

// FailureAmountMismatch marks a collection Fincra reports as paid for less
// than was asked.
const FailureAmountMismatch = "AMOUNT_MISMATCH"

// CollectionStatus maps a checkout or charge status to the platform status.
// Fincra reports an underpayment as success (and reverses it to the payer
// by default), so an underpaid collection is FAILED here; an overpayment is
// settled at the expected amount and counts as paid.
func CollectionStatus(status, varianceType string) commonv1.STATUS {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case StatusSuccess, StatusSuccessful:
		if strings.EqualFold(varianceType, VarianceUnderpayment) {
			return commonv1.STATUS_FAILED
		}
		return commonv1.STATUS_SUCCESSFUL
	case StatusFailed, StatusExpired:
		return commonv1.STATUS_FAILED
	default:
		return commonv1.STATUS_IN_PROCESS
	}
}

// PayoutStatus maps a payout status to the platform status. Fincra asks
// that only successful and failed be treated as final.
func PayoutStatus(status string) commonv1.STATUS {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case StatusSuccessful, StatusSuccess:
		return commonv1.STATUS_SUCCESSFUL
	case StatusFailed:
		return commonv1.STATUS_FAILED
	default:
		return commonv1.STATUS_IN_PROCESS
	}
}

// PaidAmount is the amount a successful charge paid toward what was asked,
// or 0 when it cannot be stated. When the customer bears the fees, Fincra's
// charge amount includes them and would not match what was asked, so no
// amount is stated; the underpayment check in CollectionStatus still holds.
func (c *Charge) PaidAmount(feeBearer string) float64 {
	if feeBearer != FeeBearerBusiness {
		return 0
	}
	if v := c.AmountExpected.Float(); v > 0 {
		return v
	}
	return c.Amount.Float()
}

// StateFor is the platform state that goes with a status.
func StateFor(status commonv1.STATUS) commonv1.STATE {
	if status == commonv1.STATUS_FAILED {
		return commonv1.STATE_INACTIVE
	}
	return commonv1.STATE_ACTIVE
}
