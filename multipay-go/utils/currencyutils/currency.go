package currencyutils

import (
	"math"

	"github.com/bojanz/currency"
)

// AmountMinorToMajor converts minor units (paisa/cents/fils) to major units (rupees/dollars/dinars)
// using the ISO 4217 minor unit exponent for the given currency.
// Examples:
//
//	AmountMinorToMajor(50000, "INR") → 500.0  (exponent 2, factor 100)
//	AmountMinorToMajor(500, "JPY")   → 500.0  (exponent 0, factor 1)
//	AmountMinorToMajor(500000, "BHD") → 500.0  (exponent 3, factor 1000)
func AmountMinorToMajor(minorAmount int64, currencyCode string) float64 {
	digits, ok := currency.GetDigits(currencyCode)
	if !ok {
		// Unknown currency — fall back to exponent 2 (most common)
		digits = 2
	}
	if digits == 0 {
		return float64(minorAmount)
	}
	factor := math.Pow(10, float64(digits))
	return float64(minorAmount) / factor
}

// AmountMajorToMinor converts major units (rupees/dollars/dinars) to minor units (paisa/cents/fils)
// using the ISO 4217 minor unit exponent for the given currency.
// Examples:
//
//	AmountMajorToMinor(500.0, "INR")  → 50000  (exponent 2, factor 100)
//	AmountMajorToMinor(500.0, "JPY")  → 500    (exponent 0, factor 1)
//	AmountMajorToMinor(500.0, "BHD")  → 500000 (exponent 3, factor 1000)
func AmountMajorToMinor(majorAmount float64, currencyCode string) int64 {
	digits, ok := currency.GetDigits(currencyCode)
	if !ok {
		// Unknown currency — fall back to exponent 2 (most common)
		digits = 2
	}
	if digits == 0 {
		return int64(math.Round(majorAmount))
	}
	factor := math.Pow(10, float64(digits))
	return int64(math.Round(majorAmount * factor))
}
