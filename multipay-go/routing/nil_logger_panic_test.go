package routing_test

import (
	"context"
	"testing"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
	"github.com/Bytonomics/multipay-india/multipay-go/providers/fake"
	"github.com/Bytonomics/multipay-india/multipay-go/routing"
)

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

func TestNewWebhookHandler_LoggerRequired(t *testing.T) {
	tests := []struct {
		name      string
		logger    ports.Logger
		wantPanic bool
	}{
		{
			name:      "nil logger panics",
			logger:    nil,
			wantPanic: true,
		},
		{
			name:      "typed nil pointer panics",
			logger:    (*ports.NoopLogger)(nil),
			wantPanic: true,
		},
		{
			name:      "real logger does not panic",
			logger:    ports.NewNoopLogger(),
			wantPanic: false,
		},
	}

	defaultHandler := func(_ context.Context, _ *domain.WebhookEvent) error { return nil }

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			construct := func() {
				routing.NewWebhookHandler(
					routing.NewEndpointMatcher("/webhooks"),
					fake.NewFakeAdapter(domain.ProviderCashfree),
					fake.NewInMemoryWebhookStore(),
					tt.logger,
					defaultHandler,
				)
			}

			if tt.wantPanic {
				panicsWithValueOrFail(t, "logger is required (cannot be nil)", construct,
					"NewWebhookHandler must panic with the exact message when the logger is nil")
				return
			}
			notPanicsOrFail(t, construct, "NewWebhookHandler must accept a non-nil logger")
		})
	}
}
