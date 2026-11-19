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

package collection

import (
	"errors"
	"fmt"
	"math"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
)

// ErrFractionalAmount is returned when a whole-units rail is asked to charge
// an amount with a fractional part.
var ErrFractionalAmount = errors.New("amount has a fractional part the rail cannot charge")

// ErrNonPositiveAmount is returned for zero or negative amounts.
var ErrNonPositiveAmount = errors.New("amount must be positive")

const nanosPerUnit = 1_000_000_000

// zeroDecimalCurrencies are ISO 4217 currencies with no minor unit.
var zeroDecimalCurrencies = map[string]bool{
	"BIF": true, "CLP": true, "DJF": true, "GNF": true, "JPY": true, "KMF": true,
	"KRW": true, "MGA": true, "PYG": true, "RWF": true, "UGX": true, "VND": true,
	"VUV": true, "XAF": true, "XOF": true, "XPF": true,
}

// threeDecimalCurrencies are ISO 4217 currencies with three minor digits.
var threeDecimalCurrencies = map[string]bool{
	"BHD": true, "IQD": true, "JOD": true, "KWD": true, "LYD": true, "OMR": true, "TND": true,
}

// CurrencyExponent is the number of minor-unit digits for a currency.
func CurrencyExponent(currency string) int {
	c := strings.ToUpper(strings.TrimSpace(currency))
	switch {
	case zeroDecimalCurrencies[c]:
		return 0
	case threeDecimalCurrencies[c]:
		return 3
	default:
		return 2
	}
}

// MinorUnits converts a Money to the smallest unit of its currency (cents,
// pesewas, …). It refuses amounts finer than the currency allows so a
// provider never charges a different amount than the one requested.
func MinorUnits(m *commonv1.Money) (int64, error) {
	if m == nil {
		return 0, ErrNonPositiveAmount
	}
	units, nanos := m.GetUnits(), int64(m.GetNanos())
	if units < 0 || nanos < 0 || (units == 0 && nanos == 0) {
		return 0, ErrNonPositiveAmount
	}
	exp := CurrencyExponent(m.GetCurrencyCode())
	scale := int64(math.Pow10(exp))
	nanosPerMinor := nanosPerUnit / scale
	if nanos%nanosPerMinor != 0 {
		return 0, fmt.Errorf("%w: %s allows %d decimals", ErrFractionalAmount, m.GetCurrencyCode(), exp)
	}
	return units*scale + nanos/nanosPerMinor, nil
}

// WholeUnits returns the amount in whole currency units, refusing fractions.
func WholeUnits(m *commonv1.Money) (int64, error) {
	if m == nil {
		return 0, ErrNonPositiveAmount
	}
	if m.GetUnits() <= 0 && m.GetNanos() <= 0 {
		return 0, ErrNonPositiveAmount
	}
	if m.GetNanos() != 0 {
		return 0, ErrFractionalAmount
	}
	return m.GetUnits(), nil
}

// DecimalString renders a Money as a plain decimal with the currency's
// minor digits ("1500.50", "1500" for UGX), the format most REST PSPs take.
func DecimalString(m *commonv1.Money) string {
	if m == nil {
		return "0"
	}
	exp := CurrencyExponent(m.GetCurrencyCode())
	if exp == 0 || m.GetNanos() == 0 {
		if exp == 0 {
			return fmt.Sprintf("%d", m.GetUnits())
		}
		return fmt.Sprintf("%d.%0*d", m.GetUnits(), exp, 0)
	}
	scale := int64(math.Pow10(exp))
	minor := int64(m.GetNanos()) / (nanosPerUnit / scale)
	return fmt.Sprintf("%d.%0*d", m.GetUnits(), exp, minor)
}

// FromMinorUnits builds a Money from a provider amount in minor units.
func FromMinorUnits(minor int64, currency string) *commonv1.Money {
	exp := CurrencyExponent(currency)
	scale := int64(math.Pow10(exp))
	return &commonv1.Money{
		CurrencyCode: strings.ToUpper(strings.TrimSpace(currency)),
		Units:        minor / scale,
		Nanos:        int32((minor % scale) * (nanosPerUnit / scale)),
	}
}
