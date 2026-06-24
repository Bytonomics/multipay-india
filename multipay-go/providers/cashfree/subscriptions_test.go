package cashfree

import (
	"testing"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

func TestBuildInlinePlanDetails_AllFields(t *testing.T) {
	pd := &domain.CreatePlanRequest{
		PlanID:         "p1",
		PlanName:       "P",
		PlanType:       domain.PlanTypePeriodic,
		Currency:       "INR",
		AmountMinor:    50000,
		MaxAmountMinor: 100000,
		Interval:       1,
		IntervalType:   domain.PlanIntervalMonth,
		MaxCycles:      12,
		Note:           "n",
	}
	d := buildInlinePlanDetails(pd)

	// All fields are pointers — deref and check non-nil first
	if d.PlanId == nil {
		t.Fatalf("PlanId is nil")
	}
	if *d.PlanId != "p1" {
		t.Fatalf("expected PlanId=p1, got %s", *d.PlanId)
	}

	if d.PlanName == nil {
		t.Fatalf("PlanName is nil")
	}
	if *d.PlanName != "P" {
		t.Fatalf("expected PlanName=P, got %s", *d.PlanName)
	}

	if d.PlanType == nil {
		t.Fatalf("PlanType is nil")
	}
	if *d.PlanType != "PERIODIC" {
		t.Fatalf("expected PlanType=PERIODIC, got %s", *d.PlanType)
	}

	if d.PlanCurrency == nil {
		t.Fatalf("PlanCurrency is nil")
	}
	if *d.PlanCurrency != "INR" {
		t.Fatalf("expected PlanCurrency=INR, got %s", *d.PlanCurrency)
	}

	if d.PlanAmount == nil {
		t.Fatalf("PlanAmount is nil")
	}
	if *d.PlanAmount != float32(500.0) {
		t.Fatalf("expected PlanAmount=500.0 (50000 minor / 100 for INR exp2), got %f", *d.PlanAmount)
	}

	if d.PlanMaxAmount == nil {
		t.Fatalf("PlanMaxAmount is nil")
	}
	if *d.PlanMaxAmount != float32(1000.0) {
		t.Fatalf("expected PlanMaxAmount=1000.0 (100000 / 100), got %f", *d.PlanMaxAmount)
	}

	if d.PlanMaxCycles == nil {
		t.Fatalf("PlanMaxCycles is nil")
	}
	if *d.PlanMaxCycles != int32(12) {
		t.Fatalf("expected PlanMaxCycles=12, got %d", *d.PlanMaxCycles)
	}

	if d.PlanIntervals == nil {
		t.Fatalf("PlanIntervals is nil")
	}
	if *d.PlanIntervals != int32(1) {
		t.Fatalf("expected PlanIntervals=1, got %d", *d.PlanIntervals)
	}

	if d.PlanIntervalType == nil {
		t.Fatalf("PlanIntervalType is nil")
	}
	if *d.PlanIntervalType != "MONTH" {
		t.Fatalf("expected PlanIntervalType=MONTH, got %s", *d.PlanIntervalType)
	}

	if d.PlanNote == nil {
		t.Fatalf("PlanNote is nil")
	}
	if *d.PlanNote != "n" {
		t.Fatalf("expected PlanNote=n, got %s", *d.PlanNote)
	}
}
