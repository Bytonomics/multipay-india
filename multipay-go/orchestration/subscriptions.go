package orchestration

import (
	"context"
	"fmt"
	"time"

	"github.com/SmrutAI/pedantigo"

	"github.com/Bytonomics/multipay-india/multipay-go/capabilities"
	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/hooks"
	"github.com/Bytonomics/multipay-india/multipay-go/logging"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
	"github.com/Bytonomics/multipay-india/multipay-go/utils/currencyutils"
)

var (
	createSubscriptionValidator      = pedantigo.New[domain.CreateSubscriptionRequest]()
	getSubscriptionValidator         = pedantigo.New[domain.GetSubscriptionRequest]()
	cancelSubscriptionValidator      = pedantigo.New[domain.CancelSubscriptionRequest]()
	pauseSubscriptionValidator       = pedantigo.New[domain.PauseSubscriptionRequest]()
	resumeSubscriptionValidator      = pedantigo.New[domain.ResumeSubscriptionRequest]()
	changePlanValidator              = pedantigo.New[domain.ChangePlanRequest]()
	getSubscriptionPaymentsValidator = pedantigo.New[domain.GetSubscriptionPaymentsRequest]()
	upgradeSubscriptionValidator     = pedantigo.New[domain.UpgradeSubscriptionRequest]()
	finalizeUpgradeValidator         = pedantigo.New[domain.FinalizeUpgradeRequest]()
	chargeSubscriptionValidator      = pedantigo.New[domain.ChargeSubscriptionRequest]()
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

// UpgradeSubscription immediately charges the pro-rata amount on an existing subscription's mandate.
// It computes the pro-rata upgrade charge and returns the charge details for the caller to finalize.
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

	prorated := domain.AmountMinor(currencyutils.ProrateUpgrade(
		int64(req.OldAmountMinor),
		int64(req.NewAmountMinor),
		req.RemainingDays,
		req.CycleDays,
		req.Currency.String(),
	))

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

	switch provider {
	case domain.ProviderCashfree:
		createReq := &domain.CreateSubscriptionRequest{
			SubscriptionID:  req.NewSubscriptionID,
			PlanID:          req.NewPlanID,
			CustomerEmail:   req.CustomerEmail,
			CustomerPhone:   req.CustomerPhone,
			CustomerName:    req.CustomerName,
			ReturnURL:       req.ReturnURL,
			FirstChargeTime: ptrTime(s.clock.Now().AddDate(0, 0, req.RemainingDays)),
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

		result = &domain.UpgradeResult{
			Strategy:                domain.UpgradeReauthProrated,
			ProratedAmountMinor:     prorated,
			RequiresReauthorization: true,
			AuthLink:                newSub.AuthLink,
			AuthSessionID:           newSub.AuthSessionID,
			Environment:             newSub.Environment,
			NewSubscriptionID:       req.NewSubscriptionID,
			RecurringEffective:      "CYCLE_END",
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
		pay, err := adapter.ChargeSubscription(ctx, &domain.ChargeSubscriptionRequest{
			SubscriptionID: req.NewSubscriptionID,
			PaymentRef:     req.PaymentRef,
			AmountMinor:    req.ProratedAmountMinor,
			Currency:       req.Currency,
			Remarks:        "plan upgrade proration",
		})
		if err != nil {
			hookCtx.Error = err
			if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
				s.logger.Error(ctx, "error in OnError hook for FinalizeUpgrade", "error", hookErr.Error())
			}
			return nil, fmt.Errorf("failed to charge upgrade proration: %w", err)
		}

		if _, cerr := adapter.CancelSubscription(ctx, &domain.CancelSubscriptionRequest{
			SubscriptionID: req.OldSubscriptionID,
		}); cerr != nil {
			hookCtx.Error = cerr
			if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, cerr); hookErr != nil {
				s.logger.Error(ctx, "error in OnError hook for FinalizeUpgrade (cancel)", "error", hookErr.Error())
			}
			return nil, fmt.Errorf("failed to cancel old subscription after upgrade: %w", cerr)
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
