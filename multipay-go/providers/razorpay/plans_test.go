package razorpay

import (
	"testing"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

func TestBuildPlanCreateData_PeriodAndItem(t *testing.T) {
	req := &domain.CreatePlanRequest{
		IntervalType: domain.PlanIntervalWeek,
		Interval:     1,
		AmountMinor:  69900,
		Currency:     "INR",
		PlanName:     "P",
	}

	data, err := buildPlanCreateData(req)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// Assert period is "weekly" (lowercase)
	if data["period"] != "weekly" {
		t.Fatalf("expected period='weekly', got %v", data["period"])
	}

	// Assert item is a nested map
	item, ok := data["item"].(map[string]any)
	if !ok {
		t.Fatalf("expected item to be map[string]any, got %T", data["item"])
	}

	// Assert item["name"] == "P"
	if item["name"] != "P" {
		t.Fatalf("expected item[name]='P', got %v", item["name"])
	}

	// Assert item["amount"] == float64(69900) (JSON numbers are float64)
	if item["amount"] != float64(69900) {
		t.Fatalf("expected item[amount]=69900.0, got %v", item["amount"])
	}

	// Assert item["currency"] == "INR"
	if item["currency"] != "INR" {
		t.Fatalf("expected item[currency]='INR', got %v", item["currency"])
	}

	// Assert period_count is NOT present
	if _, hasPeriodCount := data["period_count"]; hasPeriodCount {
		t.Fatalf("expected period_count to NOT be present, but it is")
	}
}
