package utility

import (
	"math"
	"time"

	"github.com/shopspring/decimal"
	"google.golang.org/genproto/googleapis/type/money"
)

const NanoSize = 1000000000

// maxDecimalValue returns the maximum decimal value supported
func maxDecimalValue() decimal.Decimal {
	return decimal.NewFromInt(math.MaxInt64).Add(decimal.New(999999999, -9))
}

func ToMoney(currency string, amount decimal.Decimal) money.Money {
	amount = CleanDecimal(amount)

	// Split the decimal value into units and nanos
	units := amount.IntPart()
	nanos := amount.Sub(decimal.NewFromInt(units)).Mul(decimal.NewFromInt(NanoSize)).IntPart()

	return money.Money{CurrencyCode: currency, Units: units, Nanos: int32(nanos & 0x7FFFFFFF)}
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

const (
	// DecimalPrecision defines the precision for decimal truncation.
	decimalPrecision = 9
)

func CleanDecimal(d decimal.Decimal) decimal.Decimal {
	truncatedStr := d.StringFixed(decimalPrecision)

	// Convert the string back to a decimal
	rounded, _ := decimal.NewFromString(truncatedStr)

	// Check if the value fits within the range for NUMERIC(20,9)
	// max allowed value for NUMERIC(28,9)
	minValue := maxDecimalValue().Neg() // min allowed value (negative of max)

	if rounded.GreaterThan(maxDecimalValue()) {
		return maxDecimalValue()
	} else if rounded.LessThan(minValue) {
		return minValue
	}

	return rounded
}

func IsValidTime(t *time.Time) bool {
	return t != nil && !t.IsZero()
}
