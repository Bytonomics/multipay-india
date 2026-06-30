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

func (f *fakeAdapter) MapOrderMetadata(ctx context.Context, metadata domain.Metadata) (map[string]interface{}, error) {
	return nil, nil
}

func (f *fakeAdapter) MapRefundMetadata(ctx context.Context, metadata domain.Metadata) (map[string]interface{}, error) {
	return nil, nil
}

func (f *fakeAdapter) MapPaymentLinkMetadata(ctx context.Context, metadata domain.Metadata) (map[string]interface{}, error) {
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
