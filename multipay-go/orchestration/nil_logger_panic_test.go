package orchestration_test

import (
	"testing"

	"github.com/Bytonomics/multipay-india/multipay-go/capabilities"
	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/hooks"
	"github.com/Bytonomics/multipay-india/multipay-go/orchestration"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
	"github.com/Bytonomics/multipay-india/multipay-go/providers/fake"
	"github.com/Bytonomics/multipay-india/multipay-go/routing"
)

// The panic-message contract for logger-taking service constructors is pinned here.
// Non-logger dependencies are real instances so the only nil under test is the logger.
// panicsWithValueOrFail fails the test unless fn panics with exactly want.
func panicsWithValueOrFail(t *testing.T, want string, fn func(), msg string) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("%s: expected panic %q, got none", msg, want)
			return
		}
		if r != want {
			t.Errorf("%s: panic value = %v, want %q", msg, r, want)
		}
	}()
	fn()
}

// notPanicsOrFail fails the test if fn panics.
func notPanicsOrFail(t *testing.T, fn func(), msg string) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s: unexpected panic: %v", msg, r)
		}
	}()
	fn()
}

func TestServiceConstructors_LoggerRequired(t *testing.T) {
	const wantPanic = "logger is required (cannot be nil)"

	tests := []struct {
		name      string
		construct func(logger ports.Logger)
	}{
		{
			name: "NewOrderService",
			construct: func(logger ports.Logger) {
				orchestration.NewOrderService(domain.ProviderCashfree, newFakeAdapter(), newValidator(), newPipeline(), logger, ports.NewRealClock())
			},
		},
		{
			name: "NewPaymentService",
			construct: func(logger ports.Logger) {
				orchestration.NewPaymentService(domain.ProviderCashfree, newFakeAdapter(), newValidator(), newPipeline(), logger, ports.NewRealClock())
			},
		},
		{
			name: "NewRefundService",
			construct: func(logger ports.Logger) {
				orchestration.NewRefundService(domain.ProviderCashfree, newFakeAdapter(), newValidator(), newPipeline(), logger, ports.NewRealClock())
			},
		},
		{
			name: "NewInstrumentService",
			construct: func(logger ports.Logger) {
				orchestration.NewInstrumentService(domain.ProviderCashfree, newFakeAdapter(), newValidator(), newPipeline(), logger, ports.NewRealClock())
			},
		},
		{
			name: "NewPaymentLinkService",
			construct: func(logger ports.Logger) {
				orchestration.NewPaymentLinkService(domain.ProviderCashfree, newFakeAdapter(), newValidator(), newPipeline(), logger, ports.NewRealClock())
			},
		},
		{
			name: "NewSubscriptionService",
			construct: func(logger ports.Logger) {
				orchestration.NewSubscriptionService(domain.ProviderCashfree, newFakeAdapter(), newValidator(), newPipeline(), logger, ports.NewRealClock())
			},
		},
		{
			name: "NewPlanService",
			construct: func(logger ports.Logger) {
				orchestration.NewPlanService(domain.ProviderCashfree, newFakeAdapter(), newPipeline(), logger, ports.NewRealClock())
			},
		},
		{
			name: "NewWebhookService",
			construct: func(logger ports.Logger) {
				orchestration.NewWebhookService(domain.ProviderCashfree, newFakeAdapter(), newPipeline(), fake.NewInMemoryWebhookStore(), routing.NewEndpointRegistry(), logger)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/nil logger panics", func(t *testing.T) {
			panicsWithValueOrFail(t, wantPanic, func() {
				tt.construct(nil)
			}, "constructor must panic with the exact message when the logger is nil")
		})

		t.Run(tt.name+"/typed nil pointer panics", func(t *testing.T) {
			panicsWithValueOrFail(t, wantPanic, func() {
				tt.construct((*ports.NoopLogger)(nil))
			}, "constructor must panic with the exact message when the logger is a typed nil pointer")
		})

		t.Run(tt.name+"/real logger does not panic", func(t *testing.T) {
			notPanicsOrFail(t, func() {
				tt.construct(ports.NewNoopLogger())
			}, "constructor must accept an explicitly passed logger")
		})
	}
}

func newFakeAdapter() ports.ProviderAdapter {
	return fake.NewFakeAdapter(domain.ProviderCashfree)
}

func newValidator() *capabilities.Validator {
	return capabilities.NewValidator(capabilities.NewSupportMatrix())
}

func newPipeline() *hooks.Pipeline {
	return hooks.NewPipeline(ports.NewNoopLogger())
}
