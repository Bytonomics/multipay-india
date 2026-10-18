package routing_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
	"github.com/Bytonomics/multipay-india/multipay-go/providers/fake"
	"github.com/Bytonomics/multipay-india/multipay-go/routing"
)

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
				assert.PanicsWithValue(t, "logger is required (cannot be nil)", construct,
					"NewWebhookHandler must panic with the exact message when the logger is nil")
				return
			}
			assert.NotPanics(t, construct, "NewWebhookHandler must accept a non-nil logger")
		})
	}
}
