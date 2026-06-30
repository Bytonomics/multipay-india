package currencyutils

import "testing"

func TestProrateUpgrade(t *testing.T) {
	tests := []struct {
		name      string
		oldAmount int64
		newAmount int64
		remaining int
		total     int
		currency  string
		expected  int64
	}{
		{
			name:      "upgrade mid-cycle",
			oldAmount: 10000, newAmount: 20000, remaining: 15, total: 30, currency: "INR",
			expected: 5000,
		},
		{
			name:      "downgrade returns 0",
			oldAmount: 20000, newAmount: 10000, remaining: 15, total: 30, currency: "INR",
			expected: 0,
		},
		{
			name:      "zero remaining days",
			oldAmount: 10000, newAmount: 20000, remaining: 0, total: 30, currency: "INR",
			expected: 0,
		},
		{
			name:      "full cycle remaining",
			oldAmount: 10000, newAmount: 20000, remaining: 30, total: 30, currency: "INR",
			expected: 10000,
		},
		{
			name:      "rounding to nearest",
			oldAmount: 0, newAmount: 10000, remaining: 1, total: 3, currency: "INR",
			expected: 3333,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProrateUpgrade(tt.oldAmount, tt.newAmount, tt.remaining, tt.total, tt.currency)
			if result != tt.expected {
				t.Errorf("ProrateUpgrade(%d, %d, %d, %d, %q) = %d, want %d",
					tt.oldAmount, tt.newAmount, tt.remaining, tt.total, tt.currency, result, tt.expected)
			}
		})
	}
}
