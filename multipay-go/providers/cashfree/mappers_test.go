package cashfree

import (
	"testing"

	cf "github.com/cashfree/cashfree-pg/v6"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// TestMapSubscriptionPaymentEntity_CurrencyNotHardcoded verifies that amount is converted
// with the passed currency, not hardcoded INR (Issue #5).
func TestMapSubscriptionPaymentEntity_CurrencyNotHardcoded(t *testing.T) {
	// JPY case: exp0 → no division
	amtJPY := float32(500.0)
	ej := &cf.SubscriptionPaymentEntity{
		PaymentAmount: &amtJPY,
		PaymentId:     ptrString("pp"),
	}
	pj, err := MapSubscriptionPaymentEntityToCanonical(ej, "JPY")
	if err != nil {
		t.Fatalf("failed to map payment for JPY: %v", err)
	}
	if pj == nil {
		t.Fatalf("expected non-nil payment for JPY, got nil")
	}
	if int64(pj.AmountMinor) != 500 {
		t.Fatalf("expected AmountMinor=500 for JPY (exp0), got %d", int64(pj.AmountMinor))
	}

	// BHD case: exp3 → *1000
	amtBHD := float32(1.0)
	eb := &cf.SubscriptionPaymentEntity{
		PaymentAmount: &amtBHD,
		PaymentId:     ptrString("pp"),
	}
	pb, err := MapSubscriptionPaymentEntityToCanonical(eb, "BHD")
	if err != nil {
		t.Fatalf("failed to map payment for BHD: %v", err)
	}
	if pb == nil {
		t.Fatalf("expected non-nil payment for BHD, got nil")
	}
	if int64(pb.AmountMinor) != 1000 {
		t.Fatalf("expected AmountMinor=1000 for BHD (exp3, 1.0*1000), got %d", int64(pb.AmountMinor))
	}
}

// TestMapPlanEntityToCanonical_AmountConversion verifies that Cashfree plan response
// major→minor conversion respects currency exponent.
func TestMapPlanEntityToCanonical_AmountConversion(t *testing.T) {
	// INR case: exp2 → /100
	ra := float32(500.0)
	cur := "INR"
	e := &cf.PlanEntity{
		PlanRecurringAmount: &ra,
		PlanCurrency:        &cur,
		PlanMaxAmount:       &ra,
	}
	p, err := MapPlanEntityToCanonical(e)
	if err != nil {
		t.Fatalf("failed to map plan for INR: %v", err)
	}
	if p == nil {
		t.Fatalf("expected non-nil plan for INR, got nil")
	}
	if int64(p.AmountMinor) != 50000 {
		t.Fatalf("expected AmountMinor=50000 for INR (500/0.01), got %d", int64(p.AmountMinor))
	}

	// JPY case: exp0 → no division
	ra = float32(500.0)
	cur = "JPY"
	e = &cf.PlanEntity{
		PlanRecurringAmount: &ra,
		PlanCurrency:        &cur,
		PlanMaxAmount:       &ra,
	}
	p, err = MapPlanEntityToCanonical(e)
	if err != nil {
		t.Fatalf("failed to map plan for JPY: %v", err)
	}
	if p == nil {
		t.Fatalf("expected non-nil plan for JPY, got nil")
	}
	if int64(p.AmountMinor) != 500 {
		t.Fatalf("expected AmountMinor=500 for JPY (exp0), got %d", int64(p.AmountMinor))
	}
}

// TestMapSubscriptionStatus_UnknownStatusDefault verifies that an unknown Cashfree
// subscription status is mapped to a safe canonical status (not empty/panic).
func TestMapSubscriptionStatus_UnknownStatusDefault(t *testing.T) {
	unknownStatus := "UNKNOWN_CF_STATUS"
	result := mapSubscriptionStatus(&unknownStatus)

	if result == "" {
		t.Errorf("expected non-empty canonical status for unknown input, got empty string")
	}
	if result == domain.SubscriptionStatusActive {
		t.Errorf("unknown status must NOT map to ACTIVE (would wrongly grant entitlement), got %q", result)
	}
	if result != domain.SubscriptionStatusPending {
		t.Errorf("expected unknown status to map to PENDING (safe non-entitling default), got %q", result)
	}

	// A known status still maps correctly.
	active := "ACTIVE"
	if got := mapSubscriptionStatus(&active); got != domain.SubscriptionStatusActive {
		t.Errorf("known ACTIVE must map to SubscriptionStatusActive, got %q", got)
	}
}
