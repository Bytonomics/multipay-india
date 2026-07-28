package domain

import "testing"

func TestIntervalOffset(t *testing.T) {
	tests := []struct {
		name       string
		it         PlanIntervalType
		n          int32
		wy, wm, wd int
	}{
		{"year 1", PlanIntervalYear, 1, 1, 0, 0},
		{"year 2", PlanIntervalYear, 2, 2, 0, 0},
		{"month 1", PlanIntervalMonth, 1, 0, 1, 0},
		{"month 3", PlanIntervalMonth, 3, 0, 3, 0},
		{"week 2", PlanIntervalWeek, 2, 0, 0, 14},
		{"day 10", PlanIntervalDay, 10, 0, 0, 10},
		{"empty type", PlanIntervalType(""), 5, 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y, m, d := IntervalOffset(tt.it, tt.n)
			if y != tt.wy || m != tt.wm || d != tt.wd {
				t.Errorf("IntervalOffset(%q, %d) = (%d, %d, %d); want (%d, %d, %d)",
					tt.it, tt.n, y, m, d, tt.wy, tt.wm, tt.wd)
			}
		})
	}
}
