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
	"fmt"
	"strconv"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
)

const (
	nanosPerCent     = 10_000_000
	maxDecimalDigits = 2
	int32Max         = 1<<31 - 1
)

// ParseAmount parses a positive decimal string with at most two decimal
// places into Money units and nanos.
func ParseAmount(amount string) (int64, int32, error) {
	amount = strings.TrimSpace(amount)
	if amount == "" {
		return 0, 0, errors.New("amount is required")
	}
	if strings.HasPrefix(amount, "-") {
		return 0, 0, fmt.Errorf("amount must be positive, got %q", amount)
	}
	wholePart, fracPart, _ := strings.Cut(amount, ".")
	if len(fracPart) > maxDecimalDigits {
		return 0, 0, fmt.Errorf("amount %q has more than %d decimal places", amount, maxDecimalDigits)
	}
	units, err := strconv.ParseInt(wholePart, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid amount %q", amount)
	}
	var cents int64
	if fracPart != "" {
		for len(fracPart) < maxDecimalDigits {
			fracPart += "0"
		}
		cents, err = strconv.ParseInt(fracPart, 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid amount %q", amount)
		}
	}
	if units == 0 && cents == 0 {
		return 0, 0, fmt.Errorf("amount must be positive, got %q", amount)
	}
	nanos := cents * int64(nanosPerCent)
	// cents is max 99 (two digits), so nanos is max 99 * 10_000_000 = 990_000_000, well below int32 max
	if nanos > int32Max {
		return 0, 0, errors.New("amount overflow")
	}
	//nolint:gosec // nanos is bounded by maxDecimalDigits to at most 990_000_000
	return units, int32(nanos), nil
}

// MoneyFromAmount builds a commonv1.Money from a decimal string and currency.
func MoneyFromAmount(amount, currency string) (*commonv1.Money, error) {
	units, nanos, err := ParseAmount(amount)
	if err != nil {
		return nil, err
	}
	return &commonv1.Money{CurrencyCode: currency, Units: units, Nanos: nanos}, nil
}

// FormatMoney renders Money for display, always with two decimal places.
func FormatMoney(m *commonv1.Money) string {
	cents := (int64(m.GetNanos()) + nanosPerCent/2) / nanosPerCent
	return fmt.Sprintf("%s %d.%02d", m.GetCurrencyCode(), m.GetUnits(), cents)
}

// AmountString renders Money as a bare decimal string without trailing zeros.
func AmountString(m *commonv1.Money) string {
	cents := (int64(m.GetNanos()) + nanosPerCent/2) / nanosPerCent
	if cents == 0 {
		return strconv.FormatInt(m.GetUnits(), 10)
	}
	s := fmt.Sprintf("%d.%02d", m.GetUnits(), cents)
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
}
