package logging

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
			name:      "typed nil pointer panics",
			delegate:  (*ports.NoopLogger)(nil),
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
				panicsWithValueOrFail(t, "logger is required (cannot be nil)", func() {
					NewCallerLogger(tt.delegate, 1)
				}, "NewCallerLogger must panic with the exact message when the delegate is nil")
				return
			}
			notPanicsOrFail(t, func() {
				NewCallerLogger(tt.delegate, 1)
			}, "NewCallerLogger must accept a non-nil delegate")
		})
	}
}
