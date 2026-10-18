package cashfree

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
)

func TestNewAdapter_LoggerRequired(t *testing.T) {
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
			cfg := &Config{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				Environment:  domain.EnvironmentSandbox,
				Logger:       tt.logger,
			}

			if tt.wantPanic {
				assert.PanicsWithValue(t, "cashfree adapter requires a non-nil Logger (pass NewNoopLogger only in unit tests)", func() {
					_, _ = NewAdapter(cfg)
				}, "NewAdapter must panic with the exact message when Config.Logger is nil")
				return
			}

			var err error
			assert.NotPanics(t, func() {
				_, err = NewAdapter(cfg)
			}, "NewAdapter must accept a non-nil Config.Logger")
			require.NoError(t, err, "NewAdapter must succeed with a valid config and logger")
		})
	}
}
