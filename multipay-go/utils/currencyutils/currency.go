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

// ProrateUpgrade returns the one-time top-up to charge for an immediate plan upgrade:
// (newAmountMinor - oldAmountMinor) * remainingDays / totalDays, rounded to the currency's
// smallest unit. Returns 0 for downgrades (new <= old), or when totalDays <= 0, or
// remainingDays <= 0. currencyCode is the ISO-4217 code (e.g. "INR").
func ProrateUpgrade(oldAmountMinor, newAmountMinor int64, remainingDays, totalDays int, currencyCode string) int64 {
	if newAmountMinor <= oldAmountMinor || totalDays <= 0 || remainingDays <= 0 {
		return 0
	}
	diffMinor := newAmountMinor - oldAmountMinor
	// integer math in minor units, rounded to nearest minor unit
	prorated := (diffMinor*int64(remainingDays) + int64(totalDays)/2) / int64(totalDays)
	return prorated
}
