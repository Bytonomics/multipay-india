package orchestration

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

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
		capabilityValidator := capabilities.NewValidator(capabilities.NewSupportMatrix())
		svc := NewSubscriptionService(domain.ProviderRazorpay, adapter, capabilityValidator, pipeline, logger, clock)

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
		capabilityValidator := capabilities.NewValidator(capabilities.NewSupportMatrix())
		svc := NewSubscriptionService(domain.ProviderRazorpay, adapter, capabilityValidator, pipeline, logger, clock)

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
		capabilityValidator := capabilities.NewValidator(capabilities.NewSupportMatrix())
		svc := NewSubscriptionService(domain.ProviderRazorpay, adapter, capabilityValidator, pipeline, logger, clock)

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
		capabilityValidator := capabilities.NewValidator(capabilities.NewSupportMatrix())
		svc := NewSubscriptionService(domain.ProviderRazorpay, adapter, capabilityValidator, pipeline, logger, clock)

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

// TestPreviewPlanChange tests PreviewPlanChange with Kind-based requests.
func TestPreviewPlanChange(t *testing.T) {
	logger := ports.NewNoopLogger()
	pipeline := hooks.NewPipeline(logger)
	clock := ports.NewRealClock()
	capabilityValidator := capabilities.NewValidator(capabilities.NewSupportMatrix())

	tests := []struct {
		name                 string
		req                  *domain.PlanChangePreviewRequest
		expectErr            bool
		expectChargeNow      int64
		expectProratedCredit int64
		expectNewRecurring   int64
		expectRecurringEff   string
	}{
		{
			name: "PlanChangeKindCreate",
			req: &domain.PlanChangePreviewRequest{
				Kind:           domain.PlanChangeKindCreate,
				NewAmountMinor: 160000,
			},
			expectErr:            false,
			expectChargeNow:      160000,
			expectProratedCredit: 0,
			expectNewRecurring:   160000,
			expectRecurringEff:   "IMMEDIATE",
		},
		{
			name: "PlanChangeKindUpgradeSameCycle",
			req: &domain.PlanChangePreviewRequest{
				Kind:               domain.PlanChangeKindUpgradeSameCycle,
				CurrentAmountMinor: 100000,
				NewAmountMinor:     160000,
				RemainingDays:      12,
				CurrentCycleDays:   30,
				Currency:           "INR",
			},
			expectErr:            false,
			expectChargeNow:      24000, // (160000-100000)*12/30 = 60000*12/30 = 24000
			expectProratedCredit: 0,
			expectNewRecurring:   160000,
			expectRecurringEff:   "CYCLE_END",
		},
		{
			name: "PlanChangeKindUpgradeCross",
			req: &domain.PlanChangePreviewRequest{
				Kind:               domain.PlanChangeKindUpgradeCross,
				CurrentAmountMinor: 160000,
				NewAmountMinor:     1600000,
				RemainingDays:      10,
				CurrentCycleDays:   30,
			},
			expectErr:            false,
			expectChargeNow:      1546667, // 1600000 - (160000*10/30) = 1600000 - 53333 = 1546667
			expectProratedCredit: 53333,   // (160000*10)/30 = 53333
			expectNewRecurring:   1600000,
			expectRecurringEff:   "IMMEDIATE",
		},
		{
			name: "PlanChangeKindDowngrade",
			req: &domain.PlanChangePreviewRequest{
				Kind:           domain.PlanChangeKindDowngrade,
				NewAmountMinor: 100000,
			},
			expectErr:            false,
			expectChargeNow:      0,
			expectProratedCredit: 0,
			expectNewRecurring:   100000,
			expectRecurringEff:   "CYCLE_END",
		},
		{
			name:      "nil request",
			req:       nil,
			expectErr: true,
		},
		{
			name: "unknown Kind",
			req: &domain.PlanChangePreviewRequest{
				Kind: "BOGUS",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSubscriptionService(domain.ProviderRazorpay, &fakeAdapter{}, capabilityValidator, pipeline, logger, clock)
			quote, err := svc.PreviewPlanChange(tt.req)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("expected nil error, got %v", err)
			}
			if quote == nil {
				t.Errorf("expected non-nil quote")
				return
			}

			if quote.ChargeNowMinor != tt.expectChargeNow {
				t.Errorf("ChargeNowMinor: expected %d, got %d", tt.expectChargeNow, quote.ChargeNowMinor)
			}
			if quote.ProratedCreditMinor != tt.expectProratedCredit {
				t.Errorf("ProratedCreditMinor: expected %d, got %d", tt.expectProratedCredit, quote.ProratedCreditMinor)
			}
			if quote.NewRecurringMinor != tt.expectNewRecurring {
				t.Errorf("NewRecurringMinor: expected %d, got %d", tt.expectNewRecurring, quote.NewRecurringMinor)
			}
			if quote.RecurringEffective != tt.expectRecurringEff {
				t.Errorf("RecurringEffective: expected %s, got %s", tt.expectRecurringEff, quote.RecurringEffective)
			}
		})
	}
}

// TestCreateSubscription_RecurringFieldBounds tests pedantigo tag constraints on recurring-schedule fields.
// The module-level validator (createSubscriptionValidator) enforces field bounds defined by pedantigo tags:
// RecurringAmountMinor (gte=0), RecurringInterval (gte=1), RecurringIntervalType (oneof), RecurringCurrency (iso4217).
// For wantErr=true cases, the test asserts that the adapter is NEVER invoked (validation short-circuits before dispatch).
func TestCreateSubscription_RecurringFieldBounds(t *testing.T) {
	tests := []struct {
		name            string
		recurring       domain.CreateSubscriptionRequest // base request with recurring fields set
		wantErr         bool
		adapterExpected bool // true if adapter should be called, false if validation should short-circuit
	}{
		{
			name: "valid recurring fields succeed",
			recurring: domain.CreateSubscriptionRequest{
				PlanID:                 "plan_x",
				ReturnURL:              "https://example.com/return",
				CustomerPhone:          "9876543210",
				CustomerEmail:          "test@example.com",
				FirstChargeWithMandate: true,
				RecurringAmountMinor:   49900,
				RecurringInterval:      1,
				RecurringIntervalType:  domain.PlanIntervalMonth,
				RecurringCurrency:      domain.Currency("INR"),
			},
			wantErr:         false,
			adapterExpected: true,
		},
		{
			name: "zero recurring fields succeed (omitempty)",
			recurring: domain.CreateSubscriptionRequest{
				PlanID:                 "plan_x",
				ReturnURL:              "https://example.com/return",
				CustomerPhone:          "9876543210",
				CustomerEmail:          "test@example.com",
				FirstChargeWithMandate: true,
				RecurringAmountMinor:   0,
				RecurringInterval:      0,
				RecurringIntervalType:  domain.PlanIntervalType(""),
				RecurringCurrency:      domain.Currency(""),
			},
			wantErr:         false,
			adapterExpected: true,
		},
		{
			name: "recurring_interval negative is rejected",
			recurring: domain.CreateSubscriptionRequest{
				PlanID:                 "plan_x",
				ReturnURL:              "https://example.com/return",
				CustomerPhone:          "9876543210",
				CustomerEmail:          "test@example.com",
				FirstChargeWithMandate: true,
				RecurringAmountMinor:   49900,
				RecurringInterval:      -1,
				RecurringIntervalType:  domain.PlanIntervalMonth,
				RecurringCurrency:      domain.Currency("INR"),
			},
			wantErr:         true,
			adapterExpected: false,
		},
		{
			name: "recurring_interval_type invalid enum is rejected",
			recurring: domain.CreateSubscriptionRequest{
				PlanID:                 "plan_x",
				ReturnURL:              "https://example.com/return",
				CustomerPhone:          "9876543210",
				CustomerEmail:          "test@example.com",
				FirstChargeWithMandate: true,
				RecurringAmountMinor:   49900,
				RecurringInterval:      1,
				RecurringIntervalType:  domain.PlanIntervalType("FORTNIGHT"),
				RecurringCurrency:      domain.Currency("INR"),
			},
			wantErr:         true,
			adapterExpected: false,
		},
		{
			name: "recurring_amount_minor negative is rejected",
			recurring: domain.CreateSubscriptionRequest{
				PlanID:                 "plan_x",
				ReturnURL:              "https://example.com/return",
				CustomerPhone:          "9876543210",
				CustomerEmail:          "test@example.com",
				FirstChargeWithMandate: true,
				RecurringAmountMinor:   domain.AmountMinor(-1),
				RecurringInterval:      1,
				RecurringIntervalType:  domain.PlanIntervalMonth,
				RecurringCurrency:      domain.Currency("INR"),
			},
			wantErr:         true,
			adapterExpected: false,
		},
		{
			name: "recurring_currency invalid iso4217 is rejected",
			recurring: domain.CreateSubscriptionRequest{
				PlanID:                 "plan_x",
				ReturnURL:              "https://example.com/return",
				CustomerPhone:          "9876543210",
				CustomerEmail:          "test@example.com",
				FirstChargeWithMandate: true,
				RecurringAmountMinor:   49900,
				RecurringInterval:      1,
				RecurringIntervalType:  domain.PlanIntervalMonth,
				RecurringCurrency:      domain.Currency("XX"),
			},
			wantErr:         true,
			adapterExpected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapterCalled := false
			adapter := &fakeAdapter{
				createSubscriptionFunc: func(ctx context.Context, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
					adapterCalled = true
					return &domain.Subscription{
						SubscriptionID: req.SubscriptionID,
						AuthLink:       "https://example.com/auth",
						AuthSessionID:  "session_123",
						Environment:    domain.EnvironmentSandbox,
					}, nil
				},
			}
			logger := ports.NewNoopLogger()
			pipeline := hooks.NewPipeline(logger)
			clock := ports.NewRealClock()
			capabilityValidator := capabilities.NewValidator(capabilities.NewSupportMatrix())
			svc := NewSubscriptionService(domain.ProviderRazorpay, adapter, capabilityValidator, pipeline, logger, clock)

			_, err := svc.CreateSubscription(context.Background(), &tt.recurring)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
			}

			// Assert adapter call expectations: for validation failures, adapter must NOT be called
			if tt.adapterExpected && !adapterCalled {
				t.Fatalf("adapter should have been called")
			}
			if !tt.adapterExpected && adapterCalled {
				t.Fatalf("adapter should NOT have been called (validation should short-circuit)")
			}
		})
	}
}

// TestCreateSubscription_FirstChargeWithMandate_StampsClock verifies that when the caller opts into
// first_charge_with_mandate (with first_charge_time nil, as the mutually-exclusive rule requires), the
// orchestration layer stamps req.FirstChargeTime to the clock's Now() BEFORE dispatching to the adapter.
// Adapters have no clock (library rule), so this stamp is the single source of the concrete "now".
func TestCreateSubscription_FirstChargeWithMandate_StampsClock(t *testing.T) {
	fixedTime := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	clock := &fixedClock{fixedTime: fixedTime}

	var capturedFirstChargeTime *time.Time
	adapter := &fakeAdapter{
		createSubscriptionFunc: func(ctx context.Context, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
			capturedFirstChargeTime = req.FirstChargeTime
			return &domain.Subscription{SubscriptionID: req.SubscriptionID}, nil
		},
	}

	logger := ports.NewNoopLogger()
	pipeline := hooks.NewPipeline(logger)
	capabilityValidator := capabilities.NewValidator(capabilities.NewSupportMatrix())
	svc := NewSubscriptionService(domain.ProviderRazorpay, adapter, capabilityValidator, pipeline, logger, clock)

	req := &domain.CreateSubscriptionRequest{
		PlanID:                 "plan_x",
		ReturnURL:              "https://example.com/return",
		CustomerPhone:          "9876543210",
		CustomerEmail:          "test@example.com",
		FirstChargeWithMandate: true,
		FirstChargeTime:        nil, // must be nil (mutually exclusive with the flag); orchestration stamps it
		RecurringAmountMinor:   49900,
		RecurringInterval:      1,
		RecurringIntervalType:  domain.PlanIntervalMonth,
		RecurringCurrency:      domain.Currency("INR"),
	}

	if _, err := svc.CreateSubscription(context.Background(), req); err != nil {
		t.Fatalf("CreateSubscription returned error: %v", err)
	}

	if capturedFirstChargeTime == nil {
		t.Fatal("expected orchestration to stamp req.FirstChargeTime, but the adapter received nil")
	}
	if !capturedFirstChargeTime.Equal(fixedTime) {
		t.Errorf("expected stamped FirstChargeTime == %v, got %v", fixedTime, *capturedFirstChargeTime)
	}
}

// TestUpgradeSubscription_CrossCycleIntegration verifies that UpgradeSubscription
// calls PreviewPlanChange internally and returns the correct upgrade result with cross-cycle charges.
func TestUpgradeSubscription_CrossCycleIntegration(t *testing.T) {
	tests := []struct {
		name                     string
		crossCycle               bool
		oldAmountMinor           domain.AmountMinor
		newAmountMinor           domain.AmountMinor
		remainingDays            int
		cycleDays                int
		newRecurringInterval     int32
		newRecurringIntervalType domain.PlanIntervalType
		expectRecurringEff       string
		expectProratedAmount     domain.AmountMinor
	}{
		{
			name:                     "cross-cycle upgrade",
			crossCycle:               true,
			oldAmountMinor:           160000,
			newAmountMinor:           1600000,
			remainingDays:            10,
			cycleDays:                30,
			newRecurringInterval:     1,
			newRecurringIntervalType: domain.PlanIntervalYear,
			expectRecurringEff:       "IMMEDIATE",
			expectProratedAmount:     1546667, // 1600000 - (160000*10/30) = 1600000 - 53333 = 1546667
		},
		{
			name:                     "same-cycle upgrade",
			crossCycle:               false,
			oldAmountMinor:           100000,
			newAmountMinor:           160000,
			remainingDays:            12,
			cycleDays:                30,
			newRecurringInterval:     0,
			newRecurringIntervalType: "",
			expectRecurringEff:       "CYCLE_END",
			expectProratedAmount:     24000, // (160000-100000)*12/30 = 60000*12/30 = 24000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &fakeAdapter{
				createSubscriptionFunc: func(ctx context.Context, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
					return &domain.Subscription{
						SubscriptionID: req.SubscriptionID,
						AuthLink:       "https://cashfree.com/auth",
						AuthSessionID:  "cf_session_123",
						Environment:    domain.EnvironmentSandbox,
					}, nil
				},
			}

			logger := ports.NewNoopLogger()
			pipeline := hooks.NewPipeline(logger)
			clock := ports.NewRealClock()
			capabilityValidator := capabilities.NewValidator(capabilities.NewSupportMatrix())
			svc := NewSubscriptionService(domain.ProviderCashfree, adapter, capabilityValidator, pipeline, logger, clock)

			req := &domain.UpgradeSubscriptionRequest{
				SubscriptionID:           "sub_old",
				NewSubscriptionID:        "sub_new",
				CurrentPlanID:            "plan_old",
				NewPlanID:                "plan_new",
				OldAmountMinor:           tt.oldAmountMinor,
				NewAmountMinor:           tt.newAmountMinor,
				RemainingDays:            tt.remainingDays,
				CycleDays:                tt.cycleDays,
				Currency:                 domain.Currency("INR"),
				CustomerEmail:            "test@example.com",
				CustomerPhone:            "+919876543210",
				CustomerName:             "Test User",
				ReturnURL:                "https://example.com/return",
				CrossCycle:               tt.crossCycle,
				NewRecurringInterval:     tt.newRecurringInterval,
				NewRecurringIntervalType: tt.newRecurringIntervalType,
			}

			result, err := svc.UpgradeSubscription(context.Background(), req)
			if err != nil {
				t.Fatalf("expected nil error for valid upgrade request, got %v", err)
			}
			if result == nil {
				t.Fatalf("expected non-nil upgrade result")
			}

			if result.RecurringEffective != tt.expectRecurringEff {
				t.Errorf("expected RecurringEffective=%s, got %s", tt.expectRecurringEff, result.RecurringEffective)
			}

			if result.ProratedAmountMinor != tt.expectProratedAmount {
				t.Errorf("expected ProratedAmountMinor=%d, got %d", tt.expectProratedAmount, result.ProratedAmountMinor)
			}

			// Cashfree upgrade goes through CreateSubscription (not ChangePlan),
			// so there is no ChangePlan ScheduleAt assertion here.
		})
	}
}
