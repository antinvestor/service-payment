package utility

import "github.com/pitabwire/util/decimalx"

// DecPtr returns a pointer to the given Decimal value.
func DecPtr(d decimalx.Decimal) *decimalx.Decimal { return &d }

// AbsDecimal returns the absolute value of the given Decimal.
func AbsDecimal(d decimalx.Decimal) decimalx.Decimal {
	if d.IsNegative() {
		return d.Neg()
	}
	return d
}

// MinDecimal returns the smaller of the two Decimal values.
func MinDecimal(a, b decimalx.Decimal) decimalx.Decimal {
	if a.LessThan(b) {
		return a
	}
	return b
}

// MaxDecimal returns the larger of the two Decimal values.
func MaxDecimal(a, b decimalx.Decimal) decimalx.Decimal {
	if a.GreaterThan(b) {
		return a
	}
	return b
}

// DerefOr dereferences the pointer, returning def if p is nil.
func DerefOr(p *decimalx.Decimal, def decimalx.Decimal) decimalx.Decimal {
	if p == nil {
		return def
	}
	return *p
}
