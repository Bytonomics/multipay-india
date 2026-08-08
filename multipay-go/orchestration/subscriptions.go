package orchestration

import (
	"context"
	"fmt"
	"time"

	"github.com/SmrutAI/pedantigo/v2/validator"

	"github.com/Bytonomics/multipay-india/multipay-go/capabilities"
	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/hooks"
	"github.com/Bytonomics/multipay-india/multipay-go/logging"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
	"github.com/Bytonomics/multipay-india/multipay-go/utils/currencyutils"
)

var (
	createSubscriptionValidator      = validator.New[domain.CreateSubscriptionRequest]()
	getSubscriptionValidator         = validator.New[domain.GetSubscriptionRequest]()
	cancelSubscriptionValidator      = validator.New[domain.CancelSubscriptionRequest]()
	pauseSubscriptionValidator       = validator.New[domain.PauseSubscriptionRequest]()
	resumeSubscriptionValidator      = validator.New[domain.ResumeSubscriptionRequest]()
	changePlanValidator              = validator.New[domain.ChangePlanRequest]()
	getSubscriptionPaymentsValidator = validator.New[domain.GetSubscriptionPaymentsRequest]()
	planChangePreviewValidator       = validator.New[domain.PlanChangePreviewRequest]()
	upgradeSubscriptionValidator     = validator.New[domain.UpgradeSubscriptionRequest]()
	finalizeUpgradeValidator         = validator.New[domain.FinalizeUpgradeRequest]()
	chargeSubscriptionValidator      = validator.New[domain.ChargeSubscriptionRequest]()
)

// SubscriptionService orchestrates subscription operations across multiple payment providers.
// Most subscription operations are FIRST-CLASS — both Cashfree and Razorpay support them.
// Upgrade and finalize operations require capability validation (CapSubscriptionUpgradeProration).
type SubscriptionService struct {
	adapter   ports.ProviderAdapter
	provider  domain.Provider
	validator *capabilities.Validator
	pipeline  *hooks.Pipeline
	logger    ports.Logger
	clock     ports.Clock
}

// NewSubscriptionService constructs a SubscriptionService with required dependencies.
// Logger is mandatory and will panic if nil.
func NewSubscriptionService(provider domain.Provider, adapter ports.ProviderAdapter, validator *capabilities.Validator, pipeline *hooks.Pipeline, logger ports.Logger, clock ports.Clock) *SubscriptionService {
	if logger == nil {
		panic("logger is required (cannot be nil)")
	}
	wrappedLogger := logging.NewCallerLogger(logger, 2)

	return &SubscriptionService{
		adapter:   adapter,
		provider:  provider,
		validator: validator,
		pipeline:  pipeline,
		logger:    wrappedLogger,
		clock:     clock,
	}
}

// CreateSubscription validates input, executes hooks, and creates a subscription.
// This is a first-class operation — no capability gate.
func (s *SubscriptionService) CreateSubscription(ctx context.Context, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := createSubscriptionValidator.Validate(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	// When the caller opts into charging the first period at signup, stamp a concrete "now" here so
	// each adapter maps it to its own first-charge mechanism. Adapters have no clock (library rule).
	if req.FirstChargeWithMandate {
		now := s.clock.Now()
		req.FirstChargeTime = &now
	}

	provider := s.provider
	adapter := s.adapter

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "CreateSubscription",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	result, err := adapter.CreateSubscription(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for CreateSubscription", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("create subscription failed: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("after hook failed: %w", afterErr)
	}

	return result, nil
}

// GetSubscription validates input, executes hooks, and retrieves a subscription.
// This is a first-class operation — no capability gate.
func (s *SubscriptionService) GetSubscription(ctx context.Context, req *domain.GetSubscriptionRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := getSubscriptionValidator.Validate(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	provider := s.provider
	adapter := s.adapter

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "GetSubscription",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	result, err := adapter.GetSubscription(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for GetSubscription", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("get subscription failed: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("after hook failed: %w", afterErr)
	}

	return result, nil
}

// CancelSubscription validates input, executes hooks, and cancels a subscription.
// This is a first-class operation — no capability gate.
func (s *SubscriptionService) CancelSubscription(ctx context.Context, req *domain.CancelSubscriptionRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := cancelSubscriptionValidator.Validate(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	provider := s.provider
	adapter := s.adapter

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "CancelSubscription",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	result, err := adapter.CancelSubscription(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for CancelSubscription", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("cancel subscription failed: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("after hook failed: %w", afterErr)
	}

	return result, nil
}

// PauseSubscription validates input, executes hooks, and pauses a subscription.
// This is a first-class operation — no capability gate.
func (s *SubscriptionService) PauseSubscription(ctx context.Context, req *domain.PauseSubscriptionRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := pauseSubscriptionValidator.Validate(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	provider := s.provider
	adapter := s.adapter

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "PauseSubscription",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	result, err := adapter.PauseSubscription(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for PauseSubscription", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("pause subscription failed: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("after hook failed: %w", afterErr)
	}

	return result, nil
}

// ResumeSubscription validates input, executes hooks, and resumes a subscription.
// This is a first-class operation — no capability gate.
func (s *SubscriptionService) ResumeSubscription(ctx context.Context, req *domain.ResumeSubscriptionRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := resumeSubscriptionValidator.Validate(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	provider := s.provider
	adapter := s.adapter

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "ResumeSubscription",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	result, err := adapter.ResumeSubscription(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for ResumeSubscription", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("resume subscription failed: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("after hook failed: %w", afterErr)
	}

	return result, nil
}

// ChangePlan validates input, executes hooks, and changes a subscription's plan.
// This is a first-class operation — no capability gate.
func (s *SubscriptionService) ChangePlan(ctx context.Context, req *domain.ChangePlanRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := changePlanValidator.Validate(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	provider := s.provider
	adapter := s.adapter

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "ChangePlan",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	result, err := adapter.ChangePlan(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for ChangePlan", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("change plan failed: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("after hook failed: %w", afterErr)
	}

	return result, nil
}

// GetSubscriptionPayments validates input, executes hooks, and retrieves subscription payments.
// This is a first-class operation — no capability gate.
func (s *SubscriptionService) GetSubscriptionPayments(ctx context.Context, req *domain.GetSubscriptionPaymentsRequest) ([]*domain.SubscriptionPayment, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := getSubscriptionPaymentsValidator.Validate(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	provider := s.provider
	adapter := s.adapter

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "GetSubscriptionPayments",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	result, err := adapter.GetSubscriptionPayments(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for GetSubscriptionPayments", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("get subscription payments failed: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("after hook failed: %w", afterErr)
	}

	return result, nil
}

// PreviewPlanChange computes the cost breakdown for a plan change without making any external calls.
// This is a pure function that performs no mutations and requires no capability validation.
// It validates the request and returns a quote with charge amounts and effective dates.
func (s *SubscriptionService) PreviewPlanChange(req *domain.PlanChangePreviewRequest) (*domain.PlanChangeQuote, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}
	if err := planChangePreviewValidator.Validate(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	switch req.Kind {
	case domain.PlanChangeKindCreate:
		return &domain.PlanChangeQuote{
			Kind:               domain.PlanChangeKindCreate,
			ChargeNowMinor:     req.NewAmountMinor,
			NewRecurringMinor:  req.NewAmountMinor,
			RecurringEffective: "IMMEDIATE",
		}, nil
	case domain.PlanChangeKindUpgradeSameCycle:
		charge, err := currencyutils.ProrateUpgrade(req.CurrentAmountMinor, req.NewAmountMinor, req.RemainingDays, req.CurrentCycleDays, req.Currency)
		if err != nil {
			return nil, fmt.Errorf("preview upgrade same-cycle: %w", err)
		}
		return &domain.PlanChangeQuote{
			Kind:               domain.PlanChangeKindUpgradeSameCycle,
			ChargeNowMinor:     charge,
			NewRecurringMinor:  req.NewAmountMinor,
			RecurringEffective: "CYCLE_END",
		}, nil
	case domain.PlanChangeKindUpgradeCross:
		credit, err := currencyutils.ProrateUnusedCredit(req.CurrentAmountMinor, req.CurrentCycleDays, req.RemainingDays)
		if err != nil {
			return nil, fmt.Errorf("preview upgrade cross-cycle: %w", err)
		}
		charge := req.NewAmountMinor - credit
		if charge < 0 {
			charge = 0
		}
		return &domain.PlanChangeQuote{
			Kind:                domain.PlanChangeKindUpgradeCross,
			ChargeNowMinor:      charge,
			ProratedCreditMinor: credit,
			NewRecurringMinor:   req.NewAmountMinor,
			RecurringEffective:  "IMMEDIATE",
		}, nil
	case domain.PlanChangeKindDowngrade:
		return &domain.PlanChangeQuote{
			Kind:               domain.PlanChangeKindDowngrade,
			ChargeNowMinor:     0,
			NewRecurringMinor:  req.NewAmountMinor,
			RecurringEffective: "CYCLE_END",
		}, nil
	default:
		return nil, fmt.Errorf("unknown plan change kind %q: %w", req.Kind, domain.ErrInvalidRequest)
	}
}

// UpgradeSubscription immediately charges the pro-rata amount on an existing subscription's mandate.
// It computes the pro-rata upgrade charge and returns the charge details for the caller to finalize.
// Supports both same-cycle upgrades (RecurringEffective="CYCLE_END") and cross-cycle upgrades
// (e.g., MONTHLY→YEARLY with RecurringEffective="IMMEDIATE").
// This operation requires CapSubscriptionUpgradeProration capability.
func (s *SubscriptionService) UpgradeSubscription(ctx context.Context, req *domain.UpgradeSubscriptionRequest) (*domain.UpgradeResult, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := upgradeSubscriptionValidator.Validate(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	provider := s.provider
	adapter := s.adapter

	if err := s.validator.RequireCapability(ctx, provider, capabilities.CapSubscriptionUpgradeProration); err != nil {
		return nil, fmt.Errorf("capability check failed: %w", err)
	}

	var proratedAmount int64
	if req.CrossCycle {
		credit, cerr := currencyutils.ProrateUnusedCredit(int64(req.OldAmountMinor), int64(req.CycleDays), int64(req.RemainingDays))
		if cerr != nil {
			return nil, fmt.Errorf("cross-cycle unused credit: %w", cerr)
		}
		charge := int64(req.NewAmountMinor) - credit
		if charge < 0 {
			charge = 0
		}
		proratedAmount = charge
	} else {
		p, perr := currencyutils.ProrateUpgrade(int64(req.OldAmountMinor), int64(req.NewAmountMinor), int64(req.RemainingDays), int64(req.CycleDays), req.Currency.String())
		if perr != nil {
			return nil, fmt.Errorf("upgrade proration failed: %w", perr)
		}
		proratedAmount = p
	}
	prorated := domain.AmountMinor(proratedAmount)

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "UpgradeSubscription",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	var result *domain.UpgradeResult

	var recurringEffective string

	switch provider {
	case domain.ProviderCashfree:
		// First-charge scheduling differs by upgrade kind:
		//   same-cycle  -> the delta covers the REMAINDER of the current cycle at the new tier, so the
		//                  new mandate's first auto-charge belongs at the current cycle end.
		//   cross-cycle -> the customer just paid a FULL new-cycle amount (less unused credit), so the
		//                  first auto-charge must be one NEW interval out. Scheduling it at the old
		//                  cycle end would debit another full cycle within weeks.
		firstChargeYears, firstChargeMonths, firstChargeDays := 0, 0, req.RemainingDays
		if req.CrossCycle {
			firstChargeYears, firstChargeMonths, firstChargeDays = domain.IntervalOffset(
				req.NewRecurringIntervalType, req.NewRecurringInterval)
		}

		createReq := &domain.CreateSubscriptionRequest{
			SubscriptionID:  req.NewSubscriptionID,
			PlanID:          req.NewPlanID,
			CustomerEmail:   req.CustomerEmail,
			CustomerPhone:   req.CustomerPhone,
			CustomerName:    req.CustomerName,
			ReturnURL:       req.ReturnURL,
			FirstChargeTime: ptrTime(s.clock.Now().AddDate(firstChargeYears, firstChargeMonths, firstChargeDays)),
		}
		newSub, err := adapter.CreateSubscription(ctx, createReq)
		if err != nil {
			hookCtx.Error = err
			if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
				s.logger.Error(ctx, "error in OnError hook for UpgradeSubscription", "error", hookErr.Error())
			}
			return nil, fmt.Errorf("failed to create upgrade subscription: %w", err)
		}

		if newSub == nil {
			hookCtx.Error = domain.ErrProviderError
			if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, domain.ErrProviderError); hookErr != nil {
				s.logger.Error(ctx, "error in OnError hook for UpgradeSubscription", "error", hookErr.Error())
			}
			return nil, fmt.Errorf("create subscription returned nil response: %w", domain.ErrProviderError)
		}

		// Determine recurring effective date based on cross-cycle flag
		if req.CrossCycle {
			recurringEffective = "IMMEDIATE"
		} else {
			recurringEffective = "CYCLE_END"
		}

		result = &domain.UpgradeResult{
			Strategy:                domain.UpgradeReauthProrated,
			ProratedAmountMinor:     prorated,
			RequiresReauthorization: true,
			AuthLink:                newSub.AuthLink,
			AuthSessionID:           newSub.AuthSessionID,
			Environment:             newSub.Environment,
			NewSubscriptionID:       req.NewSubscriptionID,
			RecurringEffective:      recurringEffective,
		}
	case domain.ProviderRazorpay:
		_, err := adapter.ChangePlan(ctx, &domain.ChangePlanRequest{
			SubscriptionID: req.SubscriptionID,
			NewPlanID:      req.NewPlanID,
			ScheduleAt:     domain.ScheduleChangeNow,
		})
		if err != nil {
			hookCtx.Error = err
			if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
				s.logger.Error(ctx, "error in OnError hook for UpgradeSubscription", "error", hookErr.Error())
			}
			return nil, fmt.Errorf("failed to change plan for upgrade: %w", err)
		}
		result = &domain.UpgradeResult{
			Strategy:                domain.UpgradeNativeImmediate,
			ProratedAmountMinor:     0,
			RequiresReauthorization: false,
			NewSubscriptionID:       req.NewSubscriptionID,
			RecurringEffective:      "IMMEDIATE",
		}
	default:
		return nil, fmt.Errorf("upgrade not supported for provider %s: %w", provider, domain.ErrInvalidRequest)
	}

	hookCtx.ResponseData = result
	if afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx); afterErr != nil {
		return nil, fmt.Errorf("after hook failed: %w", afterErr)
	}

	return result, nil
}

// FinalizeUpgrade persists the subscription's plan change after a successful UpgradeSubscription charge.
// This tells the adapter to flip the subscription to the new plan.
// This operation requires CapSubscriptionUpgradeProration capability.
func (s *SubscriptionService) FinalizeUpgrade(ctx context.Context, req *domain.FinalizeUpgradeRequest) (*domain.SubscriptionPayment, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := finalizeUpgradeValidator.Validate(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	provider := s.provider
	adapter := s.adapter

	if err := s.validator.RequireCapability(ctx, provider, capabilities.CapSubscriptionUpgradeProration); err != nil {
		return nil, fmt.Errorf("capability check failed: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "FinalizeUpgrade",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	var result *domain.SubscriptionPayment

	switch provider {
	case domain.ProviderCashfree:
		// When the caller already collected the prorated delta out-of-band (e.g. a one-time Order
		// settled via Orders().CreateOrder), raising a charge here would debit the customer a SECOND
		// time. In that mode we perform ONLY the provider transition: cancel the old mandate.
		var pay *domain.SubscriptionPayment
		if !req.ProrationCollectedExternally {
			charged, cerr := adapter.ChargeSubscription(ctx, &domain.ChargeSubscriptionRequest{
				SubscriptionID: req.NewSubscriptionID,
				PaymentRef:     req.PaymentRef,
				AmountMinor:    req.ProratedAmountMinor,
				Currency:       req.Currency,
				Remarks:        "plan upgrade proration",
			})
			if cerr != nil {
				hookCtx.Error = cerr
				if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, cerr); hookErr != nil {
					s.logger.Error(ctx, "error in OnError hook for FinalizeUpgrade", "error", hookErr.Error())
				}
				return nil, fmt.Errorf("failed to charge upgrade proration: %w", cerr)
			}
			pay = charged
		}

		if _, cancelErr := adapter.CancelSubscription(ctx, &domain.CancelSubscriptionRequest{
			SubscriptionID: req.OldSubscriptionID,
		}); cancelErr != nil {
			hookCtx.Error = cancelErr
			if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, cancelErr); hookErr != nil {
				s.logger.Error(ctx, "error in OnError hook for FinalizeUpgrade (cancel)", "error", hookErr.Error())
			}
			return nil, fmt.Errorf("failed to cancel old subscription after upgrade: %w", cancelErr)
		}

		if pay == nil {
			// Never return (nil, nil) for a pointer + error pair. Mirrors the Razorpay arm.
			pay = &domain.SubscriptionPayment{}
		}
		result = pay
	case domain.ProviderRazorpay:
		result = &domain.SubscriptionPayment{}
	default:
		return nil, fmt.Errorf("finalize upgrade not supported for provider %s: %w", provider, domain.ErrInvalidRequest)
	}

	hookCtx.ResponseData = result
	if afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx); afterErr != nil {
		return nil, fmt.Errorf("after hook failed: %w", afterErr)
	}

	return result, nil
}

// ChargeSubscription manually charges an existing subscription (used by upgrade, recovery, and admin operations).
// This operation requires CapSubscriptionManualCharge capability.
func (s *SubscriptionService) ChargeSubscription(ctx context.Context, req *domain.ChargeSubscriptionRequest) (*domain.SubscriptionPayment, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := chargeSubscriptionValidator.Validate(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	provider := s.provider
	adapter := s.adapter

	if err := s.validator.RequireCapability(ctx, provider, capabilities.CapSubscriptionManualCharge); err != nil {
		return nil, fmt.Errorf("capability check failed: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "ChargeSubscription",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	result, err := adapter.ChargeSubscription(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for ChargeSubscription", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("failed to charge subscription: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("after hook failed: %w", afterErr)
	}

	return result, nil
}

func ptrTime(t time.Time) *time.Time { return &t }
