package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Bytonomics/multipay-india/multipay-go/ports"
)

func TestNewCallerLogger_NilLoggerPanics(t *testing.T) {
	tests := []struct {
		name      string
		delegate  ports.Logger
		wantPanic bool
	}{
		{
			name:      "nil logger panics",
			delegate:  nil,
			wantPanic: true,
		},
		{
			name:      "real logger does not panic",
			delegate:  ports.NewNoopLogger(),
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.PanicsWithValue(t, "logging.NewCallerLogger: logger is required", func() {
					NewCallerLogger(tt.delegate, 1)
				}, "NewCallerLogger must panic with the exact message when the delegate is nil")
				return
			}
			assert.NotPanics(t, func() {
				NewCallerLogger(tt.delegate, 1)
			}, "NewCallerLogger must accept a non-nil delegate")
		})
	}
}
