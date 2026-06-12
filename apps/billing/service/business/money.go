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

package business

import (
	"errors"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/pitabwire/util/decimalx"
	utilmoney "github.com/pitabwire/util/moneyx"
)

// moneyFromDecimal converts a decimalx.Decimal to a commonv1.Money.
// Returns an error when d is nil, zero, negative, or has more than two
// decimal places (to prevent silent rounding on financial amounts).
func moneyFromDecimal(d *decimalx.Decimal, currency string) (*commonv1.Money, error) {
	if d == nil {
		return nil, errors.New("amount must not be nil")
	}
	if currency == "" {
		return nil, errors.New("currency must not be empty")
	}
	if d.IsNegative() || d.IsZero() {
		return nil, errors.New("amount must be positive")
	}

	// Reject values with more than 2 decimal places to avoid silent rounding.
	// Convert to cents (minor units at 2dp) and back; mismatch means precision was lost.
	const centsScale = 2
	cents := d.ToMinorUnits(centsScale)
	reconstructed := decimalx.FromMinorUnits(cents, centsScale)
	if reconstructed.Cmp(*d) != 0 {
		return nil, errors.New("amount has more than 2 decimal places: silent rounding not allowed")
	}

	return utilmoney.ToMoney(currency, *d), nil
}

// moneyMatchesDecimal returns true when the Money value equals the decimal
// amount and currency.  Used for amount-match checks in SettleFromCheckout.
func moneyMatchesDecimal(m *commonv1.Money, d *decimalx.Decimal, currency string) bool {
	if m == nil || d == nil {
		return false
	}
	if m.GetCurrencyCode() != currency {
		return false
	}
	fromMoney := utilmoney.FromMoney(m)
	return fromMoney.Cmp(*d) == 0
}
