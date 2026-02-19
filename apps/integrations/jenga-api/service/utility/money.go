package utility

import (
	"math"
	"time"

	"github.com/shopspring/decimal"
	"google.golang.org/genproto/googleapis/type/money"
)

const (
	NanoSize      = 1000000000
	decimalPlaces = 9
	maxNanoSize   = 999999999 // Maximum nanos value for money.Money
)

// getMaxDecimalValue returns the maximum allowed decimal value for NUMERIC(28,9).
func getMaxDecimalValue() decimal.Decimal {
	return decimal.NewFromInt(math.MaxInt64).Add(decimal.New(maxNanoSize, -9))
}

func ToMoney(currency string, amount decimal.Decimal) money.Money {
	amount = CleanDecimal(amount)

	// Split the decimal value into units and nanos
	units := amount.IntPart()
	nanos := amount.Sub(decimal.NewFromInt(units)).Mul(decimal.NewFromInt(NanoSize)).IntPart()

	return money.Money{
		CurrencyCode: currency,
		Units:        units,
		Nanos:        int32(nanos), //nolint:gosec // G115: Safe conversion, nanos value is always less than 1 billion
	}
}

func FromMoney(m *money.Money) decimal.Decimal {
	units := decimal.NewFromInt(m.GetUnits())
	nanos := decimal.NewFromInt(int64(m.GetNanos())).Div(decimal.NewFromInt(NanoSize))
	return units.Add(nanos)
}

func CompareMoney(a, b *money.Money) bool {
	if a.GetCurrencyCode() != b.GetCurrencyCode() {
		return false
	}
	if a.GetUnits() != b.GetUnits() {
		return false
	}
	if a.GetNanos() != b.GetNanos() {
		return false
	}
	return true
}

func CleanDecimal(d decimal.Decimal) decimal.Decimal {
	truncatedStr := d.StringFixed(decimalPlaces)

	// Convert the string back to a decimal
	rounded, _ := decimal.NewFromString(truncatedStr)

	// Check if the value fits within the range for NUMERIC(20,9)
	// max allowed value for NUMERIC(28,9)
	maxValue := getMaxDecimalValue()
	minValue := maxValue.Neg() // min allowed value (negative of max)

	if rounded.GreaterThan(maxValue) {
		return maxValue
	} else if rounded.LessThan(minValue) {
		return minValue
	}

	return rounded
}

func IsValidTime(t *time.Time) bool {
	return t != nil && !t.IsZero()
}
