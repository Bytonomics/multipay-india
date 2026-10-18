package client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Bytonomics/multipay-india/multipay-go/client"
	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
	"github.com/Bytonomics/multipay-india/multipay-go/providers/fake"
)

func TestNewClient_LoggerRequired(t *testing.T) {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &client.ClientConfig{
				Provider:     fake.NewFakeAdapter(domain.ProviderCashfree),
				WebhookStore: fake.NewInMemoryWebhookStore(),
				Logger:       tt.logger,
			}

			if tt.wantPanic {
				assert.PanicsWithValue(t, "client.NewClient: logger is required", func() {
					_, _ = client.NewClient(cfg)
				}, "NewClient must panic with the exact message when ClientConfig.Logger is nil")
				return
			}

			var err error
			assert.NotPanics(t, func() {
				_, err = client.NewClient(cfg)
			}, "NewClient must accept a non-nil ClientConfig.Logger")
			require.NoError(t, err, "NewClient must succeed with a valid config and logger")
		})
	}
}
