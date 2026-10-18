package fake

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Bytonomics/multipay-india/multipay-go/ports"
)

func TestNewHarness_LoggerRequired(t *testing.T) {
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
			if tt.wantPanic {
				assert.PanicsWithValue(t, "fake.NewHarness: logger is required", func() {
					NewHarness(&Config{Logger: tt.logger})
				}, "NewHarness must panic with the exact message when Config.Logger is nil")
				return
			}
			assert.NotPanics(t, func() {
				NewHarness(&Config{Logger: tt.logger})
			}, "NewHarness must accept a non-nil Config.Logger")
		})
	}
}
