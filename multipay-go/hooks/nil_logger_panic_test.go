package hooks

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

func TestConstructors_NilLoggerPanics(t *testing.T) {
	tests := []struct {
		name      string
		construct func(logger ports.Logger)
		wantPanic string
	}{
		{
			name: "NewPipeline",
			construct: func(logger ports.Logger) {
				NewPipeline(logger)
			},
			wantPanic: "logger is required (cannot be nil)",
		},
		{
			name: "NewAuditHook",
			construct: func(logger ports.Logger) {
				NewAuditHook(logger)
			},
			wantPanic: "logger is required (cannot be nil)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/nil logger panics", func(t *testing.T) {
			panicsWithValueOrFail(t, tt.wantPanic, func() {
				tt.construct(nil)
			}, "constructor must panic with the exact message when the logger is nil")
		})

		t.Run(tt.name+"/typed nil pointer panics", func(t *testing.T) {
			panicsWithValueOrFail(t, tt.wantPanic, func() {
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
