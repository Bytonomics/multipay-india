package fake

import (
	"testing"

	"github.com/Bytonomics/multipay-india/multipay-go/ports"
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				panicsWithValueOrFail(t, "logger is required (cannot be nil)", func() {
					NewHarness(&Config{Logger: tt.logger})
				}, "NewHarness must panic with the exact message when Config.Logger is nil")
				return
			}
			notPanicsOrFail(t, func() {
				NewHarness(&Config{Logger: tt.logger})
			}, "NewHarness must accept a non-nil Config.Logger")
		})
	}
}
