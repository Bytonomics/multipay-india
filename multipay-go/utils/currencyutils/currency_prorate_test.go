package currencyutils

import (
	"errors"
	"math"
	"testing"
)

func TestProrateUpgrade(t *testing.T) {
	tests := []struct {
		name             string
		oldAmount        int64
		newAmount        int64
		remaining        int64
		total            int64
		currency         string
		expectedProrated int64
		expectedErr      error
		wantOverflow     bool
	}{
		// Normal upgrade cases
		{
			name:             "upgrade mid-cycle",
			oldAmount:        10000,
			newAmount:        20000,
			remaining:        15,
			total:            30,
			currency:         "INR",
			expectedProrated: 5000,
			expectedErr:      nil,
		},
		{
			name:             "downgrade returns 0",
			oldAmount:        20000,
			newAmount:        10000,
			remaining:        15,
			total:            30,
			currency:         "INR",
			expectedProrated: 0,
			expectedErr:      nil,
		},
		{
			name:             "zero remaining days",
			oldAmount:        10000,
			newAmount:        20000,
			remaining:        0,
			total:            30,
			currency:         "INR",
			expectedProrated: 0,
			expectedErr:      nil,
		},
		{
			name:             "full cycle remaining",
			oldAmount:        10000,
			newAmount:        20000,
			remaining:        30,
			total:            30,
			currency:         "INR",
			expectedProrated: 10000,
			expectedErr:      nil,
		},
		{
			name:             "rounding to nearest",
			oldAmount:        0,
			newAmount:        10000,
			remaining:        1,
			total:            3,
			currency:         "INR",
			expectedProrated: 3333,
			expectedErr:      nil,
		},
		{
			name:             "zero total days",
			oldAmount:        10000,
			newAmount:        20000,
			remaining:        15,
			total:            0,
			currency:         "INR",
			expectedProrated: 0,
			expectedErr:      nil,
		},
		{
			name:             "equal old and new amounts",
			oldAmount:        10000,
			newAmount:        10000,
			remaining:        15,
			total:            30,
			currency:         "INR",
			expectedProrated: 0,
			expectedErr:      nil,
		},
		// Overflow case: charge * days exceeds math.MaxInt64
		{
			name:             "overflow: charge * remaining days exceeds int64 max",
			oldAmount:        0,
			newAmount:        math.MaxInt64,
			remaining:        1000,
			total:            1,
			currency:         "INR",
			expectedProrated: 0,
			expectedErr:      ErrProrationOverflow,
			wantOverflow:     true,
		},
		{
			name:             "overflow via rounding addend (product fits, +totalDays/2 overflows)",
			oldAmount:        0,
			newAmount:        math.MaxInt64,
			remaining:        1,
			total:            2,
			currency:         "INR",
			expectedProrated: 0,
			expectedErr:      ErrProrationOverflow,
			wantOverflow:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProrateUpgrade(tt.oldAmount, tt.newAmount, tt.remaining, tt.total, tt.currency)

			if tt.wantOverflow {
				if !errors.Is(err, ErrProrationOverflow) {
					t.Errorf("ProrateUpgrade(%d, %d, %d, %d, %q) error = %v, want %v",
						tt.oldAmount, tt.newAmount, tt.remaining, tt.total, tt.currency, err, ErrProrationOverflow)
				}
				if result != 0 {
					t.Errorf("ProrateUpgrade(%d, %d, %d, %d, %q) result = %d, want 0 on overflow",
						tt.oldAmount, tt.newAmount, tt.remaining, tt.total, tt.currency, result)
				}
				return
			}

			if err != tt.expectedErr {
				t.Errorf("ProrateUpgrade(%d, %d, %d, %d, %q) error = %v, want %v",
					tt.oldAmount, tt.newAmount, tt.remaining, tt.total, tt.currency, err, tt.expectedErr)
			}
			if result != tt.expectedProrated {
				t.Errorf("ProrateUpgrade(%d, %d, %d, %d, %q) = %d, want %d",
					tt.oldAmount, tt.newAmount, tt.remaining, tt.total, tt.currency, result, tt.expectedProrated)
			}
		})
	}
}

func TestProrateUnusedCredit(t *testing.T) {
	tests := []struct {
		name          string
		paidAmount    int64
		cycleDays     int64
		remainingDays int64
		expected      int64
	}{
		// Basic calculation: 10000 paid for 30 days, 15 days remain = 5000
		{
			name:          "basic credit calculation",
			paidAmount:    10000,
			cycleDays:     30,
			remainingDays: 15,
			expected:      5000, // (10000 * 15) / 30 = 5000
		},
		// All days remain: 10000 paid for 30 days, all 30 days remain = 10000
		{
			name:          "all days remaining",
			paidAmount:    10000,
			cycleDays:     30,
			remainingDays: 30,
			expected:      10000, // (10000 * 30) / 30 = 10000
		},
		// Zero days remain: 10000 paid for 30 days, 0 days remain = 0
		{
			name:          "zero days remaining",
			paidAmount:    10000,
			cycleDays:     30,
			remainingDays: 0,
			expected:      0, // (10000 * 0) / 30 = 0
		},
		// remainingDays exceeds cycleDays (clamped to cycleDays): 10000 paid for 30 days, 50 remain (clamp to 30) = 10000
		{
			name:          "remaining days exceeds cycle days (clamped)",
			paidAmount:    10000,
			cycleDays:     30,
			remainingDays: 50,
			expected:      10000, // Clamped to 30: (10000 * 30) / 30 = 10000
		},
		// Negative remainingDays (clamped to 0): 10000 paid for 30 days, -5 remain (clamp to 0) = 0
		{
			name:          "negative remaining days (clamped to 0)",
			paidAmount:    10000,
			cycleDays:     30,
			remainingDays: -5,
			expected:      0, // Clamped to 0: (10000 * 0) / 30 = 0
		},
		// Zero paid amount: 0 paid for 30 days, 15 days remain = 0
		{
			name:          "zero paid amount",
			paidAmount:    0,
			cycleDays:     30,
			remainingDays: 15,
			expected:      0, // (0 * 15) / 30 = 0
		},
		// Zero cycle days (returns 0): 10000 paid for 0 days, 15 days remain = 0
		{
			name:          "zero cycle days",
			paidAmount:    10000,
			cycleDays:     0,
			remainingDays: 15,
			expected:      0, // cycleDays <= 0: returns 0
		},
		// Negative cycle days (returns 0): 10000 paid for -30 days, 15 days remain = 0
		{
			name:          "negative cycle days",
			paidAmount:    10000,
			cycleDays:     -30,
			remainingDays: 15,
			expected:      0, // cycleDays <= 0: returns 0
		},
		// Large paidAmount with fractional days (rounding): 100000 paid for 30 days, 10 days remain = 33333
		{
			name:          "large amount with rounding",
			paidAmount:    100000,
			cycleDays:     30,
			remainingDays: 10,
			expected:      33333, // (100000 * 10) / 30 = 33333.33... → 33333
		},
		// Large paidAmount whose product with remainingDays still fits in int64 (no overflow).
		{
			name:          "large product within int64 (no overflow)",
			paidAmount:    9_000_000_000_000_000, // 9e15; product 9e15*999 ≈ 8.99e18 < math.MaxInt64
			cycleDays:     1000,
			remainingDays: 999,
			expected:      9_000_000_000_000_000 * 999 / 1000, // 8_991_000_000_000_000, no overflow
		},
		// Monthly to yearly cycle transition: 50000 paid for 30 days, 30 days remain (used to calculate full unused credit)
		{
			name:          "monthly to yearly: full cycle unused",
			paidAmount:    50000,
			cycleDays:     30,
			remainingDays: 30,
			expected:      50000, // Full credit: (50000 * 30) / 30 = 50000
		},
		// Monthly to yearly cycle transition: 50000 paid for 30 days, 20 days remain (partial unused)
		{
			name:          "monthly to yearly: partial cycle unused",
			paidAmount:    50000,
			cycleDays:     30,
			remainingDays: 20,
			expected:      33333, // Partial credit: (50000 * 20) / 30 = 33333.33... → 33333
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProrateUnusedCredit(tt.paidAmount, tt.cycleDays, tt.remainingDays)
			if result != tt.expected {
				t.Errorf("ProrateUnusedCredit(%d, %d, %d) = %d, want %d",
					tt.paidAmount, tt.cycleDays, tt.remainingDays, result, tt.expected)
			}
		})
	}
}
