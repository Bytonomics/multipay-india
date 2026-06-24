package orchestration

import (
	"context"
	"fmt"

	"github.com/SmrutAI/pedantigo"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/hooks"
	"github.com/Bytonomics/multipay-india/multipay-go/logging"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
)

var (
	createSubscriptionValidator      = pedantigo.New[domain.CreateSubscriptionRequest]()
	getSubscriptionValidator         = pedantigo.New[domain.GetSubscriptionRequest]()
	cancelSubscriptionValidator      = pedantigo.New[domain.CancelSubscriptionRequest]()
	pauseSubscriptionValidator       = pedantigo.New[domain.PauseSubscriptionRequest]()
	resumeSubscriptionValidator      = pedantigo.New[domain.ResumeSubscriptionRequest]()
	changePlanValidator              = pedantigo.New[domain.ChangePlanRequest]()
	getSubscriptionPaymentsValidator = pedantigo.New[domain.GetSubscriptionPaymentsRequest]()
)

// SubscriptionService orchestrates subscription operations across multiple payment providers.
// Subscriptions are FIRST-CLASS — both Cashfree and Razorpay support all 7 operations.
// No capability validation is required (unlike capability-gated services).
type SubscriptionService struct {
	adapter  ports.ProviderAdapter
	provider domain.Provider
	pipeline *hooks.Pipeline
	logger   ports.Logger
	clock    ports.Clock
}

// NewSubscriptionService constructs a SubscriptionService with required dependencies.
// Logger is mandatory and will panic if nil.
func NewSubscriptionService(provider domain.Provider, adapter ports.ProviderAdapter, pipeline *hooks.Pipeline, logger ports.Logger, clock ports.Clock) *SubscriptionService {
	if logger == nil {
		panic("logger is required (cannot be nil)")
	}
	wrappedLogger := logging.NewCallerLogger(logger, 2)

	return &SubscriptionService{
		adapter:  adapter,
		provider: provider,
		pipeline: pipeline,
		logger:   wrappedLogger,
		clock:    clock,
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
