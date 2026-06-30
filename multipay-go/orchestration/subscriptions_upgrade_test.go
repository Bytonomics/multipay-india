package orchestration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Bytonomics/multipay-india/multipay-go/capabilities"
	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/hooks"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
	"github.com/Bytonomics/multipay-india/multipay-go/utils/currencyutils"
)

// fixedClock is a deterministic Clock implementation for testing.
// It returns a fixed time so assertions on FirstChargeTime are exact.
type fixedClock struct {
	fixedTime time.Time
}

func (c *fixedClock) Now() time.Time {
	return c.fixedTime
}

// TestUpgradeSubscription_Cashfree_CreatesReauthSubscription tests Cashfree upgrade path:
// Creates a new subscription with FirstChargeTime deferred by RemainingDays,
// returns UpgradeReauthProrated strategy, requires reauthorization.
func TestUpgradeSubscription_Cashfree_CreatesReauthSubscription(t *testing.T) {
	// Fixed clock for deterministic FirstChargeTime assertion
	fixedTime := time.Date(2026, 6, 25, 10, 0, 0, 0, time.UTC)
	clock := &fixedClock{fixedTime: fixedTime}

	createSubCalled := false
	var capturedCreateReq *domain.CreateSubscriptionRequest
	expectedAuthLink := "https://auth.cashfree.com/sub_123"

	adapter := &fakeAdapter{
		createSubscriptionFunc: func(ctx context.Context, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
			createSubCalled = true
			capturedCreateReq = req
			return &domain.Subscription{
				SubscriptionID: req.SubscriptionID,
				AuthLink:       expectedAuthLink,
			}, nil
		},
	}

	logger := ports.NewNoopLogger()
	pipeline := hooks.NewPipeline(logger)
	validator := capabilities.NewValidator(capabilities.NewSupportMatrix())
	svc := NewSubscriptionService(domain.ProviderCashfree, adapter, validator, pipeline, logger, clock)

	req := &domain.UpgradeSubscriptionRequest{
		SubscriptionID:    "sub_old",
		NewSubscriptionID: "sub_new",
		CurrentPlanID:     "plan_old",
		NewPlanID:         "plan_new",
		OldAmountMinor:    domain.AmountMinor(50000),
		NewAmountMinor:    domain.AmountMinor(75000),
		Currency:          "INR",
		RemainingDays:     10,
		CycleDays:         30,
		CustomerEmail:     "user@example.com",
		CustomerPhone:     "+919876543210",
		CustomerName:      "John Doe",
		ReturnURL:         "https://example.com/return",
	}

	result, err := svc.UpgradeSubscription(context.Background(), req)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// Assert adapter.CreateSubscription was called
	if !createSubCalled {
		t.Fatalf("expected adapter.CreateSubscription to be called for Cashfree")
	}

	// Assert CreateSubscriptionRequest has correct fields
	if capturedCreateReq.SubscriptionID != req.NewSubscriptionID {
		t.Errorf("expected SubscriptionID=%s, got %s", req.NewSubscriptionID, capturedCreateReq.SubscriptionID)
	}
	if capturedCreateReq.PlanID != req.NewPlanID {
		t.Errorf("expected PlanID=%s, got %s", req.NewPlanID, capturedCreateReq.PlanID)
	}

	// Assert FirstChargeTime is correctly set to clock.Now() + RemainingDays
	expectedFirstChargeTime := clock.Now().AddDate(0, 0, req.RemainingDays)
	if capturedCreateReq.FirstChargeTime == nil {
		t.Fatalf("expected non-nil FirstChargeTime")
	}
	// Compare date components (Year, Month, Day) to avoid timezone/nanosecond issues
	if capturedCreateReq.FirstChargeTime.Year() != expectedFirstChargeTime.Year() ||
		capturedCreateReq.FirstChargeTime.Month() != expectedFirstChargeTime.Month() ||
		capturedCreateReq.FirstChargeTime.Day() != expectedFirstChargeTime.Day() {
		t.Errorf("expected FirstChargeTime date=%v, got %v",
			expectedFirstChargeTime.Format("2006-01-02"),
			capturedCreateReq.FirstChargeTime.Format("2006-01-02"))
	}

	// Assert UpgradeResult fields
	if result.Strategy != domain.UpgradeReauthProrated {
		t.Errorf("expected Strategy=REAUTH_PRORATED, got %s", result.Strategy)
	}
	if !result.RequiresReauthorization {
		t.Errorf("expected RequiresReauthorization=true, got %v", result.RequiresReauthorization)
	}
	if result.AuthLink != expectedAuthLink {
		t.Errorf("expected AuthLink=%s, got %s", expectedAuthLink, result.AuthLink)
	}
	if result.NewSubscriptionID != req.NewSubscriptionID {
		t.Errorf("expected NewSubscriptionID=%s, got %s", req.NewSubscriptionID, result.NewSubscriptionID)
	}
	if result.RecurringEffective != "CYCLE_END" {
		t.Errorf("expected RecurringEffective=CYCLE_END, got %s", result.RecurringEffective)
	}

	// Assert prorated amount is correctly calculated
	expectedProrated := domain.AmountMinor(currencyutils.ProrateUpgrade(
		int64(req.OldAmountMinor),
		int64(req.NewAmountMinor),
		req.RemainingDays,
		req.CycleDays,
		req.Currency.String(),
	))
	if result.ProratedAmountMinor != expectedProrated {
		t.Errorf("expected ProratedAmountMinor=%d, got %d", expectedProrated, result.ProratedAmountMinor)
	}
}

// TestUpgradeSubscription_Razorpay_ChangesPlainImmediate tests Razorpay upgrade path:
// Calls ChangePlan with ScheduleAt=NOW, returns UpgradeNativeImmediate strategy,
// no reauthorization needed.
func TestUpgradeSubscription_Razorpay_ChangesPlainImmediate(t *testing.T) {
	fixedTime := time.Date(2026, 6, 25, 10, 0, 0, 0, time.UTC)
	clock := &fixedClock{fixedTime: fixedTime}

	changePlanCalled := false
	var capturedChangePlanReq *domain.ChangePlanRequest

	adapter := &fakeAdapter{
		changePlanFunc: func(ctx context.Context, req *domain.ChangePlanRequest) (*domain.Subscription, error) {
			changePlanCalled = true
			capturedChangePlanReq = req
			return &domain.Subscription{
				SubscriptionID: req.SubscriptionID,
			}, nil
		},
	}

	logger := ports.NewNoopLogger()
	pipeline := hooks.NewPipeline(logger)
	validator := capabilities.NewValidator(capabilities.NewSupportMatrix())
	svc := NewSubscriptionService(domain.ProviderRazorpay, adapter, validator, pipeline, logger, clock)

	req := &domain.UpgradeSubscriptionRequest{
		SubscriptionID:    "sub_old",
		NewSubscriptionID: "sub_new",
		CurrentPlanID:     "plan_old",
		NewPlanID:         "plan_new",
		OldAmountMinor:    domain.AmountMinor(50000),
		NewAmountMinor:    domain.AmountMinor(75000),
		Currency:          "INR",
		RemainingDays:     10,
		CycleDays:         30,
		CustomerEmail:     "user@example.com",
		CustomerPhone:     "+919876543210",
		CustomerName:      "John Doe",
		ReturnURL:         "https://example.com/return",
	}

	result, err := svc.UpgradeSubscription(context.Background(), req)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// Assert adapter.ChangePlan was called
	if !changePlanCalled {
		t.Fatalf("expected adapter.ChangePlan to be called for Razorpay")
	}

	// Assert ChangePlanRequest has correct fields
	if capturedChangePlanReq.SubscriptionID != req.SubscriptionID {
		t.Errorf("expected SubscriptionID=%s, got %s", req.SubscriptionID, capturedChangePlanReq.SubscriptionID)
	}
	if capturedChangePlanReq.NewPlanID != req.NewPlanID {
		t.Errorf("expected NewPlanID=%s, got %s", req.NewPlanID, capturedChangePlanReq.NewPlanID)
	}
	if capturedChangePlanReq.ScheduleAt != domain.ScheduleChangeNow {
		t.Errorf("expected ScheduleAt=NOW, got %s", capturedChangePlanReq.ScheduleAt)
	}

	// Assert UpgradeResult fields
	if result.Strategy != domain.UpgradeNativeImmediate {
		t.Errorf("expected Strategy=NATIVE_IMMEDIATE, got %s", result.Strategy)
	}
	if result.RequiresReauthorization {
		t.Errorf("expected RequiresReauthorization=false, got %v", result.RequiresReauthorization)
	}
	if result.ProratedAmountMinor != 0 {
		t.Errorf("expected ProratedAmountMinor=0 for Razorpay, got %d", result.ProratedAmountMinor)
	}
	if result.NewSubscriptionID != req.NewSubscriptionID {
		t.Errorf("expected NewSubscriptionID=%s, got %s", req.NewSubscriptionID, result.NewSubscriptionID)
	}
	if result.RecurringEffective != "IMMEDIATE" {
		t.Errorf("expected RecurringEffective=IMMEDIATE, got %s", result.RecurringEffective)
	}
}

// TestUpgradeSubscription_NilRequest tests nil request handling.
func TestUpgradeSubscription_NilRequest(t *testing.T) {
	clock := &fixedClock{fixedTime: time.Now()}
	adapter := &fakeAdapter{}
	logger := ports.NewNoopLogger()
	pipeline := hooks.NewPipeline(logger)
	validator := capabilities.NewValidator(capabilities.NewSupportMatrix())
	svc := NewSubscriptionService(domain.ProviderCashfree, adapter, validator, pipeline, logger, clock)

	_, err := svc.UpgradeSubscription(context.Background(), nil)
	if err == nil {
		t.Fatalf("expected error for nil request, got nil")
	}
	if !errors.Is(err, domain.ErrInvalidRequest) {
		t.Errorf("expected error to wrap ErrInvalidRequest, got %v", err)
	}
}

// TestFinalizeUpgrade_Cashfree_ChargesAndCancels tests Cashfree finalize path:
// Charges the new subscription for the prorated amount, then cancels the old subscription.
func TestFinalizeUpgrade_Cashfree_ChargesAndCancels(t *testing.T) {
	fixedTime := time.Date(2026, 6, 25, 10, 0, 0, 0, time.UTC)
	clock := &fixedClock{fixedTime: fixedTime}

	chargeCalled := false
	cancelCalled := false
	var capturedChargeReq *domain.ChargeSubscriptionRequest
	var capturedCancelReq *domain.CancelSubscriptionRequest
	expectedPaymentID := "pay_12345"

	adapter := &fakeAdapter{
		chargeSubscriptionFunc: func(ctx context.Context, req *domain.ChargeSubscriptionRequest) (*domain.SubscriptionPayment, error) {
			chargeCalled = true
			capturedChargeReq = req
			return &domain.SubscriptionPayment{
				PaymentID:   expectedPaymentID,
				Status:      domain.SubPaymentStatusSuccess,
				AmountMinor: req.AmountMinor,
			}, nil
		},
		cancelSubscriptionFunc: func(ctx context.Context, req *domain.CancelSubscriptionRequest) (*domain.Subscription, error) {
			cancelCalled = true
			capturedCancelReq = req
			return &domain.Subscription{}, nil
		},
	}

	logger := ports.NewNoopLogger()
	pipeline := hooks.NewPipeline(logger)
	validator := capabilities.NewValidator(capabilities.NewSupportMatrix())
	svc := NewSubscriptionService(domain.ProviderCashfree, adapter, validator, pipeline, logger, clock)

	req := &domain.FinalizeUpgradeRequest{
		NewSubscriptionID:   "sub_new",
		OldSubscriptionID:   "sub_old",
		PaymentRef:          "pay_ref_123",
		ProratedAmountMinor: domain.AmountMinor(12500),
		Currency:            "INR",
	}

	payment, err := svc.FinalizeUpgrade(context.Background(), req)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// Assert adapter.ChargeSubscription was called
	if !chargeCalled {
		t.Fatalf("expected adapter.ChargeSubscription to be called for Cashfree")
	}
	if capturedChargeReq.SubscriptionID != req.NewSubscriptionID {
		t.Errorf("expected ChargeSubscription SubscriptionID=%s, got %s", req.NewSubscriptionID, capturedChargeReq.SubscriptionID)
	}
	if capturedChargeReq.PaymentRef != req.PaymentRef {
		t.Errorf("expected PaymentRef=%s, got %s", req.PaymentRef, capturedChargeReq.PaymentRef)
	}
	if capturedChargeReq.AmountMinor != req.ProratedAmountMinor {
		t.Errorf("expected AmountMinor=%d, got %d", req.ProratedAmountMinor, capturedChargeReq.AmountMinor)
	}

	// Assert adapter.CancelSubscription was called
	if !cancelCalled {
		t.Fatalf("expected adapter.CancelSubscription to be called for Cashfree")
	}
	if capturedCancelReq.SubscriptionID != req.OldSubscriptionID {
		t.Errorf("expected CancelSubscription SubscriptionID=%s, got %s", req.OldSubscriptionID, capturedCancelReq.SubscriptionID)
	}

	// Assert returned payment
	if payment == nil {
		t.Fatalf("expected non-nil payment, got nil")
	}
	if payment.PaymentID != expectedPaymentID {
		t.Errorf("expected PaymentID=%s, got %s", expectedPaymentID, payment.PaymentID)
	}
}

// TestFinalizeUpgrade_Razorpay_ReturnsEmptyPayment tests Razorpay finalize path:
// Returns empty SubscriptionPayment with no adapter calls.
func TestFinalizeUpgrade_Razorpay_ReturnsEmptyPayment(t *testing.T) {
	fixedTime := time.Date(2026, 6, 25, 10, 0, 0, 0, time.UTC)
	clock := &fixedClock{fixedTime: fixedTime}

	chargeCalled := false
	cancelCalled := false

	adapter := &fakeAdapter{
		chargeSubscriptionFunc: func(ctx context.Context, req *domain.ChargeSubscriptionRequest) (*domain.SubscriptionPayment, error) {
			chargeCalled = true
			return &domain.SubscriptionPayment{}, nil
		},
		cancelSubscriptionFunc: func(ctx context.Context, req *domain.CancelSubscriptionRequest) (*domain.Subscription, error) {
			cancelCalled = true
			return &domain.Subscription{}, nil
		},
	}

	logger := ports.NewNoopLogger()
	pipeline := hooks.NewPipeline(logger)
	validator := capabilities.NewValidator(capabilities.NewSupportMatrix())
	svc := NewSubscriptionService(domain.ProviderRazorpay, adapter, validator, pipeline, logger, clock)

	req := &domain.FinalizeUpgradeRequest{
		NewSubscriptionID:   "sub_new",
		OldSubscriptionID:   "sub_old",
		PaymentRef:          "pay_ref_123",
		ProratedAmountMinor: domain.AmountMinor(20000),
		Currency:            "INR",
	}

	payment, err := svc.FinalizeUpgrade(context.Background(), req)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// Assert adapter methods were NOT called for Razorpay
	if chargeCalled {
		t.Fatalf("expected adapter.ChargeSubscription NOT to be called for Razorpay")
	}
	if cancelCalled {
		t.Fatalf("expected adapter.CancelSubscription NOT to be called for Razorpay")
	}

	// Assert returned payment is non-nil
	if payment == nil {
		t.Fatalf("expected non-nil payment, got nil")
	}
}

// TestFinalizeUpgrade_Cashfree_ChargeError tests error handling when charge fails.
func TestFinalizeUpgrade_Cashfree_ChargeError(t *testing.T) {
	fixedTime := time.Date(2026, 6, 25, 10, 0, 0, 0, time.UTC)
	clock := &fixedClock{fixedTime: fixedTime}

	cancelCalled := false
	expectedErr := errors.New("charge failed: insufficient funds")

	adapter := &fakeAdapter{
		chargeSubscriptionFunc: func(ctx context.Context, req *domain.ChargeSubscriptionRequest) (*domain.SubscriptionPayment, error) {
			return nil, expectedErr
		},
		cancelSubscriptionFunc: func(ctx context.Context, req *domain.CancelSubscriptionRequest) (*domain.Subscription, error) {
			cancelCalled = true
			return &domain.Subscription{}, nil
		},
	}

	logger := ports.NewNoopLogger()
	pipeline := hooks.NewPipeline(logger)
	validator := capabilities.NewValidator(capabilities.NewSupportMatrix())
	svc := NewSubscriptionService(domain.ProviderCashfree, adapter, validator, pipeline, logger, clock)

	req := &domain.FinalizeUpgradeRequest{
		NewSubscriptionID:   "sub_new",
		OldSubscriptionID:   "sub_old",
		PaymentRef:          "pay_ref_123",
		ProratedAmountMinor: domain.AmountMinor(12500),
		Currency:            "INR",
	}

	payment, err := svc.FinalizeUpgrade(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error for charge failure, got nil")
	}

	// Assert payment is nil
	if payment != nil {
		t.Fatalf("expected nil payment on error, got %v", payment)
	}

	// Assert CancelSubscription was NOT called when charge fails
	if cancelCalled {
		t.Fatalf("expected adapter.CancelSubscription NOT to be called when charge fails")
	}
}
