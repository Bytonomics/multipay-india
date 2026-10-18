package hooks

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Bytonomics/multipay-india/multipay-go/ports"
)

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
			wantPanic: "hooks.NewPipeline: logger is required",
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
			assert.PanicsWithValue(t, tt.wantPanic, func() {
				tt.construct(nil)
			}, "constructor must panic with the exact message when the logger is nil")
		})

		t.Run(tt.name+"/real logger does not panic", func(t *testing.T) {
			assert.NotPanics(t, func() {
				tt.construct(ports.NewNoopLogger())
			}, "constructor must accept an explicitly passed logger")
		})
	}
}
