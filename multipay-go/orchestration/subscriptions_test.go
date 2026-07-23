package orchestration

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Bytonomics/multipay-india/multipay-go/capabilities"
	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/hooks"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
)

// fakeAdapter is a test implementation of ports.ProviderAdapter.
// It implements all 29 methods of the ProviderAdapter interface.
// Operation methods can be configured with custom func fields; all other methods return zero values.
type fakeAdapter struct {
	createPlanFunc         func(ctx context.Context, req *domain.CreatePlanRequest) (*domain.Plan, error)
	getPlanFunc            func(ctx context.Context, req *domain.GetPlanRequest) (*domain.Plan, error)
	createSubscriptionFunc func(ctx context.Context, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error)
	changePlanFunc         func(ctx context.Context, req *domain.ChangePlanRequest) (*domain.Subscription, error)
	cancelSubscriptionFunc func(ctx context.Context, req *domain.CancelSubscriptionRequest) (*domain.Subscription, error)
	chargeSubscriptionFunc func(ctx context.Context, req *domain.ChargeSubscriptionRequest) (*domain.SubscriptionPayment, error)
}

// --- OrderProvider methods ---

func (f *fakeAdapter) CreateOrder(ctx context.Context, req *domain.CreateOrderRequest) (*domain.Order, error) {
	return nil, nil
}

func (f *fakeAdapter) GetOrder(ctx context.Context, req *domain.GetOrderRequest) (*domain.Order, error) {
	return nil, nil
}

func (f *fakeAdapter) ListOrderPayments(ctx context.Context, req *domain.ListOrderPaymentsRequest) ([]*domain.Payment, error) {
	return nil, nil
}

// --- PaymentProvider methods ---

func (f *fakeAdapter) GetPayment(ctx context.Context, req *domain.GetPaymentRequest) (*domain.Payment, error) {
	return nil, nil
}

func (f *fakeAdapter) ListPayments(ctx context.Context, req *domain.ListPaymentsRequest) ([]*domain.Payment, error) {
	return nil, nil
}

func (f *fakeAdapter) CapturePayment(ctx context.Context, req *domain.CapturePaymentRequest) (*domain.Payment, error) {
	return nil, nil
}

// --- RefundProvider methods ---

func (f *fakeAdapter) CreateRefund(ctx context.Context, req *domain.CreateRefundRequest) (*domain.Refund, error) {
	return nil, nil
}

func (f *fakeAdapter) GetRefund(ctx context.Context, req *domain.GetRefundRequest) (*domain.Refund, error) {
	return nil, nil
}

func (f *fakeAdapter) ListRefunds(ctx context.Context, req *domain.ListRefundsRequest) ([]*domain.Refund, error) {
	return nil, nil
}

// --- InstrumentProvider methods ---

func (f *fakeAdapter) GetInstrument(ctx context.Context, req *domain.GetInstrumentRequest) (*domain.Instrument, error) {
	return nil, nil
}

func (f *fakeAdapter) ListInstruments(ctx context.Context, req *domain.ListInstrumentsRequest) ([]*domain.Instrument, error) {
	return nil, nil
}

func (f *fakeAdapter) DeleteInstrument(ctx context.Context, req *domain.DeleteInstrumentRequest) (*domain.Instrument, error) {
	return nil, nil
}

// --- PaymentLinkProvider methods ---

func (f *fakeAdapter) CreatePaymentLink(ctx context.Context, req *domain.CreatePaymentLinkRequest) (*domain.PaymentLink, error) {
	return nil, nil
}

func (f *fakeAdapter) GetPaymentLink(ctx context.Context, req *domain.GetPaymentLinkRequest) (*domain.PaymentLink, error) {
	return nil, nil
}

func (f *fakeAdapter) CancelPaymentLink(ctx context.Context, req *domain.CancelPaymentLinkRequest) (*domain.PaymentLink, error) {
	return nil, nil
}

// --- PlanProvider methods ---

func (f *fakeAdapter) CreatePlan(ctx context.Context, req *domain.CreatePlanRequest) (*domain.Plan, error) {
	if f.createPlanFunc != nil {
		return f.createPlanFunc(ctx, req)
	}
	return nil, nil
}

func (f *fakeAdapter) GetPlan(ctx context.Context, req *domain.GetPlanRequest) (*domain.Plan, error) {
	if f.getPlanFunc != nil {
		return f.getPlanFunc(ctx, req)
	}
	return nil, nil
}

// --- SubscriptionProvider methods ---

func (f *fakeAdapter) CreateSubscription(ctx context.Context, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
	if f.createSubscriptionFunc != nil {
		return f.createSubscriptionFunc(ctx, req)
	}
	return nil, nil
}

func (f *fakeAdapter) GetSubscription(ctx context.Context, req *domain.GetSubscriptionRequest) (*domain.Subscription, error) {
	return nil, nil
}

func (f *fakeAdapter) CancelSubscription(ctx context.Context, req *domain.CancelSubscriptionRequest) (*domain.Subscription, error) {
	if f.cancelSubscriptionFunc != nil {
		return f.cancelSubscriptionFunc(ctx, req)
	}
	return &domain.Subscription{}, nil
}

func (f *fakeAdapter) PauseSubscription(ctx context.Context, req *domain.PauseSubscriptionRequest) (*domain.Subscription, error) {
	return nil, nil
}

func (f *fakeAdapter) ResumeSubscription(ctx context.Context, req *domain.ResumeSubscriptionRequest) (*domain.Subscription, error) {
	return nil, nil
}

func (f *fakeAdapter) ChangePlan(ctx context.Context, req *domain.ChangePlanRequest) (*domain.Subscription, error) {
	if f.changePlanFunc != nil {
		return f.changePlanFunc(ctx, req)
	}
	return &domain.Subscription{}, nil
}

func (f *fakeAdapter) GetSubscriptionPayments(ctx context.Context, req *domain.GetSubscriptionPaymentsRequest) ([]*domain.SubscriptionPayment, error) {
	return nil, nil
}

func (f *fakeAdapter) ChargeSubscription(ctx context.Context, req *domain.ChargeSubscriptionRequest) (*domain.SubscriptionPayment, error) {
	if f.chargeSubscriptionFunc != nil {
		return f.chargeSubscriptionFunc(ctx, req)
	}
	return &domain.SubscriptionPayment{
		PaymentID:   req.PaymentRef,
		Status:      domain.SubPaymentStatusSuccess,
		AmountMinor: req.AmountMinor,
	}, nil
}

// --- WebhookConsumerProvider methods ---

func (f *fakeAdapter) VerifySignature(ctx context.Context, payload []byte, headers map[string]string) error {
	return nil
}

func (f *fakeAdapter) ParseEvent(ctx context.Context, payload []byte, headers map[string]string) (*domain.WebhookEvent, error) {
	return &domain.WebhookEvent{}, nil
}

func (f *fakeAdapter) SupportedWebhookEvents() []domain.WebhookEventType {
	return nil
}

// --- MetadataMapper methods ---

func (f *fakeAdapter) MapOrderMetadata(ctx context.Context, metadata domain.Metadata) (map[string]any, error) {
	return nil, nil
}

func (f *fakeAdapter) MapRefundMetadata(ctx context.Context, metadata domain.Metadata) (map[string]any, error) {
	return nil, nil
}

func (f *fakeAdapter) MapPaymentLinkMetadata(ctx context.Context, metadata domain.Metadata) (map[string]any, error) {
	return nil, nil
}

// --- Adapter identification ---

func (f *fakeAdapter) ProviderName() domain.Provider {
	return domain.ProviderRazorpay
}

func (f *fakeAdapter) ProviderCapabilities() []capabilities.Capability {
	return nil
}

// --- Test cases ---

// TestCreateSubscriptionRequest_Validate_XOR tests the XOR validation of PlanID and PlanDetails.
func TestCreateSubscriptionRequest_Validate_XOR(t *testing.T) {
	base := &domain.CreateSubscriptionRequest{
		SubscriptionID: "s",
		CustomerEmail:  "a@b.com",
		CustomerPhone:  "12345",
		ReturnURL:      "https://example.com/return",
	}

	// Case A: neither PlanID nor PlanDetails set → error
	{
		req := *base
		if err := createSubscriptionValidator.Validate(&req); err == nil {
			t.Fatalf("expected error for neither plan_id nor plan_details, got nil")
		}
	}

	// Case B: both PlanID and PlanDetails set → error
	{
		req := *base
		req.PlanID = "p1"
		req.PlanDetails = &domain.CreatePlanRequest{
			PlanID:       "p2",
			PlanName:     "P",
			PlanType:     domain.PlanTypePeriodic,
			Currency:     "INR",
			AmountMinor:  50000,
			Interval:     1,
			IntervalType: domain.PlanIntervalMonth,
		}
		if err := createSubscriptionValidator.Validate(&req); err == nil {
			t.Fatalf("expected error for both plan_id and plan_details, got nil")
		}
	}

	// Case C: only PlanID set → nil
	{
		req := *base
		req.PlanID = "p1"
		if err := createSubscriptionValidator.Validate(&req); err != nil {
			t.Fatalf("expected nil for only plan_id, got %v", err)
		}
	}

	// Case D: only PlanDetails set (valid PERIODIC) → nil
	{
		req := *base
		req.PlanDetails = &domain.CreatePlanRequest{
			PlanID:         "p2",
			PlanName:       "P",
			PlanType:       domain.PlanTypePeriodic,
			MaxAmountMinor: 100000,
			Currency:       "INR",
			AmountMinor:    50000,
			Interval:       1,
			IntervalType:   domain.PlanIntervalMonth,
		}
		if err := createSubscriptionValidator.Validate(&req); err != nil {
			t.Fatalf("expected nil for only plan_details, got %v", err)
		}
	}

	// Case E: empty CustomerEmail (optional at domain level — Cashfree enforces it, for razorpay it's optional) → nil
	{
		req := *base
		req.CustomerEmail = ""
		req.PlanID = "p1"
		if err := createSubscriptionValidator.Validate(&req); err != nil {
			t.Fatalf("expected nil for empty customer_email (optional at domain level), got %v", err)
		}
	}
}

// TestChangePlanRequest_Validation tests the validation of ScheduleAt enum.
func TestChangePlanRequest_Validation(t *testing.T) {
	base := &domain.ChangePlanRequest{
		SubscriptionID: "s",
		NewPlanID:      "p2",
	}

	// Case A: ScheduleAt="LATER" → error (oneof=NOW,CYCLE_END)
	{
		req := *base
		req.ScheduleAt = "LATER"
		if err := changePlanValidator.Validate(&req); err == nil {
			t.Fatalf("expected error for ScheduleAt=LATER, got nil")
		}
	}

	// Case B: ScheduleAt=domain.ScheduleChangeNow → nil
	{
		req := *base
		req.ScheduleAt = domain.ScheduleChangeNow
		if err := changePlanValidator.Validate(&req); err != nil {
			t.Fatalf("expected nil for ScheduleAt=NOW, got %v", err)
		}
	}

	// Case C: ScheduleAt=domain.ScheduleChangeCycleEnd → nil
	{
		req := *base
		req.ScheduleAt = domain.ScheduleChangeCycleEnd
		if err := changePlanValidator.Validate(&req); err != nil {
			t.Fatalf("expected nil for ScheduleAt=CYCLE_END, got %v", err)
		}
	}
}

// TestSubscriptionService_CreateSubscription_Pipeline tests the request validation and adapter call pipeline.
func TestSubscriptionService_CreateSubscription_Pipeline(t *testing.T) {
	// Case A: nil req → error, adapter NOT called
	{
		adapterCalled := false
		adapter := &fakeAdapter{
			createSubscriptionFunc: func(ctx context.Context, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
				adapterCalled = true
				return nil, nil
			},
		}
		logger := ports.NewNoopLogger()
		pipeline := hooks.NewPipeline(logger)
		clock := ports.NewRealClock()
		validator := capabilities.NewValidator(capabilities.NewSupportMatrix())
		svc := NewSubscriptionService(domain.ProviderRazorpay, adapter, validator, pipeline, logger, clock)

		_, err := svc.CreateSubscription(context.Background(), nil)
		if err == nil {
			t.Fatalf("expected error for nil req, got nil")
		}
		if adapterCalled {
			t.Fatalf("adapter should NOT have been called for nil req")
		}
	}

	// Case B: invalid req (neither plan_id nor plan_details) → XOR error, adapter NOT called
	{
		adapterCalled := false
		adapter := &fakeAdapter{
			createSubscriptionFunc: func(ctx context.Context, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
				adapterCalled = true
				return &domain.Subscription{}, nil
			},
		}
		logger := ports.NewNoopLogger()
		pipeline := hooks.NewPipeline(logger)
		clock := ports.NewRealClock()
		validator := capabilities.NewValidator(capabilities.NewSupportMatrix())
		svc := NewSubscriptionService(domain.ProviderRazorpay, adapter, validator, pipeline, logger, clock)

		req := &domain.CreateSubscriptionRequest{SubscriptionID: "s", CustomerEmail: "a@b.com", CustomerPhone: "12345", ReturnURL: "https://example.com/return"}
		// missing both PlanID and PlanDetails — XOR violation caught by Validate()
		_, err := svc.CreateSubscription(context.Background(), req)
		if err == nil {
			t.Fatalf("expected error for invalid req, got nil")
		}
		if adapterCalled {
			t.Fatalf("adapter should NOT have been called for invalid req")
		}
	}

	// Case C: valid req, adapter returns subscription → returns it, nil error
	{
		expectedSub := &domain.Subscription{SubscriptionID: "x"}
		adapter := &fakeAdapter{
			createSubscriptionFunc: func(ctx context.Context, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
				return expectedSub, nil
			},
		}
		logger := ports.NewNoopLogger()
		pipeline := hooks.NewPipeline(logger)
		clock := ports.NewRealClock()
		validator := capabilities.NewValidator(capabilities.NewSupportMatrix())
		svc := NewSubscriptionService(domain.ProviderRazorpay, adapter, validator, pipeline, logger, clock)

		req := &domain.CreateSubscriptionRequest{
			SubscriptionID: "s",
			CustomerEmail:  "a@b.com",
			CustomerPhone:  "12345",
			PlanID:         "p",
			ReturnURL:      "https://example.com/return",
		}
		sub, err := svc.CreateSubscription(context.Background(), req)
		if err != nil {
			t.Fatalf("expected nil error for valid req, got %v", err)
		}
		if sub != expectedSub {
			t.Fatalf("expected returned subscription to match, got %v", sub)
		}
	}

	// Case D: adapter returns error → error is wrapped/contains "boom"
	{
		adapter := &fakeAdapter{
			createSubscriptionFunc: func(ctx context.Context, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
				return nil, errors.New("boom")
			},
		}
		logger := ports.NewNoopLogger()
		pipeline := hooks.NewPipeline(logger)
		clock := ports.NewRealClock()
		validator := capabilities.NewValidator(capabilities.NewSupportMatrix())
		svc := NewSubscriptionService(domain.ProviderRazorpay, adapter, validator, pipeline, logger, clock)

		req := &domain.CreateSubscriptionRequest{
			SubscriptionID: "s",
			CustomerEmail:  "a@b.com",
			CustomerPhone:  "12345",
			PlanID:         "p",
			ReturnURL:      "https://example.com/return",
		}
		_, err := svc.CreateSubscription(context.Background(), req)
		if err == nil {
			t.Fatalf("expected error for adapter boom, got nil")
		}
		if !strings.Contains(err.Error(), "boom") {
			t.Fatalf("expected error to contain 'boom', got %v", err)
		}
	}
}

// TestPreviewPlanChange_SameCycle tests PreviewPlanChange for an upgrade within MONTHLY cycle.
// Expects chargeNow to be prorated based on remainingDays, and RecurringEffective to be CYCLE_END.
func TestPreviewPlanChange_SameCycle(t *testing.T) {
	req := &domain.PlanChangePreviewRequest{
		PlanKey:       "plan_monthly",
		CurrentAmt:    50000, // ₹500 in minor units
		NewAmt:        75000, // ₹750 in minor units (upgrade)
		RemainingDays: 15,
		CycleDays:     30,
		CycleType:     "MONTHLY",
	}

	logger := ports.NewNoopLogger()
	pipeline := hooks.NewPipeline(logger)
	clock := ports.NewRealClock()
	validator := capabilities.NewValidator(capabilities.NewSupportMatrix())
	svc := NewSubscriptionService(domain.ProviderRazorpay, &fakeAdapter{}, validator, pipeline, logger, clock)

	quote, err := svc.PreviewPlanChange(req)
	if err != nil {
		t.Fatalf("expected nil error for valid same-cycle request, got %v", err)
	}
	if quote == nil {
		t.Fatalf("expected non-nil quote for same-cycle request")
	}

	if quote.RecurringEffective != "CYCLE_END" {
		t.Errorf("expected RecurringEffective=CYCLE_END for same-cycle, got %s", quote.RecurringEffective)
	}

	if quote.ChargeNowMinor == 0 {
		t.Errorf("expected non-zero chargeNow for same-cycle upgrade, got %d", quote.ChargeNowMinor)
	}

	if quote.Kind != domain.PlanChangeKindUpgrade {
		t.Errorf("expected Kind=PlanChangeKindUpgrade, got %v", quote.Kind)
	}
}

// TestPreviewPlanChange_CrossCycle tests PreviewPlanChange for an upgrade from MONTHLY to YEARLY.
// Expects chargeNow = annual - unused-credit, and RecurringEffective to be IMMEDIATE.
func TestPreviewPlanChange_CrossCycle(t *testing.T) {
	req := &domain.PlanChangePreviewRequest{
		PlanKey:       "plan_yearly",
		CurrentAmt:    50000,  // ₹500/month in minor units
		NewAmt:        600000, // ₹6000 annual in minor units
		RemainingDays: 15,
		CycleDays:     30,
		CycleType:     "YEARLY", // upgrading from MONTHLY to YEARLY
	}

	logger := ports.NewNoopLogger()
	pipeline := hooks.NewPipeline(logger)
	clock := ports.NewRealClock()
	validator := capabilities.NewValidator(capabilities.NewSupportMatrix())
	svc := NewSubscriptionService(domain.ProviderRazorpay, &fakeAdapter{}, validator, pipeline, logger, clock)

	quote, err := svc.PreviewPlanChange(req)
	if err != nil {
		t.Fatalf("expected nil error for valid cross-cycle request, got %v", err)
	}
	if quote == nil {
		t.Fatalf("expected non-nil quote for cross-cycle request")
	}

	if quote.RecurringEffective != "IMMEDIATE" {
		t.Errorf("expected RecurringEffective=IMMEDIATE for cross-cycle, got %s", quote.RecurringEffective)
	}

	if quote.ChargeNowMinor == 0 {
		t.Errorf("expected non-zero chargeNow for cross-cycle upgrade, got %d", quote.ChargeNowMinor)
	}

	if quote.Kind != domain.PlanChangeKindUpgradeCross {
		t.Errorf("expected Kind=PlanChangeKindUpgradeCross, got %v", quote.Kind)
	}

	if !quote.RequiresReauthorization {
		t.Errorf("expected RequiresReauthorization=true for cross-cycle, got %v", quote.RequiresReauthorization)
	}

	// Verify exact chargeNow calculation: newAmt - prorateUnusedCredit(oldAmt, cycleDays, remainingDays)
	// = 600000 - (50000 * 15 / 30) = 600000 - 25000 = 575000
	expectedChargeNow := req.NewAmt - (req.CurrentAmt*req.RemainingDays)/req.CycleDays
	if quote.ChargeNowMinor != expectedChargeNow {
		t.Errorf("expected chargeNow=%d (600000 - 25000), got %d", expectedChargeNow, quote.ChargeNowMinor)
	}

	if quote.NewRecurringMinor != req.NewAmt {
		t.Errorf("expected NewRecurringMinor=%d (req.NewAmt), got %d", req.NewAmt, quote.NewRecurringMinor)
	}
}

// TestPreviewPlanChange_CrossCycleClampsNegativeCharge tests PreviewPlanChange for a downgrade
// from YEARLY where unused credit exceeds the new annual price, ensuring chargeNow is clamped to 0.
func TestPreviewPlanChange_CrossCycleClampsNegativeCharge(t *testing.T) {
	req := &domain.PlanChangePreviewRequest{
		PlanKey:       "plan_yearly_downgrade",
		CurrentAmt:    600000, // ₹6000/year in minor units
		NewAmt:        100000, // ₹1000/year in minor units (downgrade)
		RemainingDays: 30,
		CycleDays:     30,
		CycleType:     "YEARLY", // downgrading within YEARLY
	}

	logger := ports.NewNoopLogger()
	pipeline := hooks.NewPipeline(logger)
	clock := ports.NewRealClock()
	validator := capabilities.NewValidator(capabilities.NewSupportMatrix())
	svc := NewSubscriptionService(domain.ProviderRazorpay, &fakeAdapter{}, validator, pipeline, logger, clock)

	quote, err := svc.PreviewPlanChange(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quote == nil {
		t.Fatalf("expected non-nil quote")
	}
	if quote.ChargeNowMinor != 0 {
		t.Errorf("expected ChargeNowMinor clamped to 0 when unused credit exceeds new annual, got %d", quote.ChargeNowMinor)
	}
}

// TestUpgradeSubscription_CrossCycleIntegration verifies that UpgradeSubscription
// calls PreviewPlanChange internally and returns the correct upgrade result.
func TestUpgradeSubscription_CrossCycleIntegration(t *testing.T) {
	changeReq := (*domain.ChangePlanRequest)(nil)
	changeReqCaptured := false

	adapter := &fakeAdapter{
		changePlanFunc: func(ctx context.Context, req *domain.ChangePlanRequest) (*domain.Subscription, error) {
			changeReqCaptured = true
			changeReq = req
			return &domain.Subscription{SubscriptionID: "sub_123"}, nil
		},
	}

	logger := ports.NewNoopLogger()
	pipeline := hooks.NewPipeline(logger)
	clock := ports.NewRealClock()
	validator := capabilities.NewValidator(capabilities.NewSupportMatrix())
	svc := NewSubscriptionService(domain.ProviderRazorpay, adapter, validator, pipeline, logger, clock)

	req := &domain.UpgradeSubscriptionRequest{
		SubscriptionID:    "sub_old",
		NewSubscriptionID: "sub_new",
		CurrentPlanID:     "plan_old",
		NewPlanID:         "plan_new",
		OldAmountMinor:    50000,
		NewAmountMinor:    500000, // annual
		RemainingDays:     15,
		CycleDays:         30,
		Currency:          domain.Currency("INR"),
		CustomerEmail:     "test@example.com",
		CustomerPhone:     "+919876543210",
		CustomerName:      "Test User",
		ReturnURL:         "https://example.com/return",
		CrossCycle:        true,
	}

	result, err := svc.UpgradeSubscription(context.Background(), req)
	if err != nil {
		t.Fatalf("expected nil error for valid upgrade request, got %v", err)
	}
	if result == nil {
		t.Fatalf("expected non-nil upgrade result")
	}

	if result.RecurringEffective != "IMMEDIATE" {
		t.Errorf("expected RecurringEffective=IMMEDIATE for cross-cycle upgrade, got %s", result.RecurringEffective)
	}

	if !changeReqCaptured {
		t.Errorf("expected ChangePlan to be called for Razorpay cross-cycle upgrade")
	}

	if changeReq != nil && changeReq.ScheduleAt != domain.ScheduleChangeNow {
		t.Errorf("expected ChangePlan ScheduleAt=NOW, got %s", changeReq.ScheduleAt)
	}
}
