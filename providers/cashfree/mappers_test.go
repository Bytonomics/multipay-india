package cashfree

import (
	"testing"

	cf "github.com/cashfree/cashfree-pg/v6"
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
	pj := MapSubscriptionPaymentEntityToCanonical(ej, "JPY")
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
	pb := MapSubscriptionPaymentEntityToCanonical(eb, "BHD")
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
	p := MapPlanEntityToCanonical(e)
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
	p = MapPlanEntityToCanonical(e)
	if p == nil {
		t.Fatalf("expected non-nil plan for JPY, got nil")
	}
	if int64(p.AmountMinor) != 500 {
		t.Fatalf("expected AmountMinor=500 for JPY (exp0), got %d", int64(p.AmountMinor))
	}
}
