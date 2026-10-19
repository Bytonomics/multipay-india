package ports_test

import (
	"context"
	"testing"

	"github.com/Bytonomics/multipay-india/multipay-go/ports"
)

// testPtrLogger implements ports.Logger with pointer receivers, so only *testPtrLogger is a Logger.
type testPtrLogger struct{}

func (l *testPtrLogger) Info(_ context.Context, _ string, _ ...any)  {}
func (l *testPtrLogger) Error(_ context.Context, _ string, _ ...any) {}
func (l *testPtrLogger) Debug(_ context.Context, _ string, _ ...any) {}

// testValueLogger implements ports.Logger with value receivers, so the value itself is a Logger.
type testValueLogger struct{}

func (testValueLogger) Info(_ context.Context, _ string, _ ...any)  {}
func (testValueLogger) Error(_ context.Context, _ string, _ ...any) {}
func (testValueLogger) Debug(_ context.Context, _ string, _ ...any) {}

// testMapLogger is a map-kinded Logger, used to exercise the typed nil map branch.
type testMapLogger map[string]string

func (testMapLogger) Info(_ context.Context, _ string, _ ...any)  {}
func (testMapLogger) Error(_ context.Context, _ string, _ ...any) {}
func (testMapLogger) Debug(_ context.Context, _ string, _ ...any) {}

// testFuncLogger is a func-kinded Logger, used to exercise the typed nil func branch.
type testFuncLogger func()

func (testFuncLogger) Info(_ context.Context, _ string, _ ...any)  {}
func (testFuncLogger) Error(_ context.Context, _ string, _ ...any) {}
func (testFuncLogger) Debug(_ context.Context, _ string, _ ...any) {}

// testChanLogger is a chan-kinded Logger, used to exercise the typed nil chan branch.
type testChanLogger chan int

func (testChanLogger) Info(_ context.Context, _ string, _ ...any)  {}
func (testChanLogger) Error(_ context.Context, _ string, _ ...any) {}
func (testChanLogger) Debug(_ context.Context, _ string, _ ...any) {}

// testSliceLogger is a slice-kinded Logger, used to exercise the typed nil slice branch and the
// present non-nil empty slice case.
type testSliceLogger []string

func (testSliceLogger) Info(_ context.Context, _ string, _ ...any)  {}
func (testSliceLogger) Error(_ context.Context, _ string, _ ...any) {}
func (testSliceLogger) Debug(_ context.Context, _ string, _ ...any) {}

func TestIsNilLogger(t *testing.T) {
	var typedNilPtr *testPtrLogger
	var typedNilMap testMapLogger
	var typedNilFunc testFuncLogger
	var typedNilChan testChanLogger
	var typedNilSlice testSliceLogger

	tests := []struct {
		name   string
		logger ports.Logger
		want   bool
	}{
		{
			name:   "untyped nil interface is missing",
			logger: nil,
			want:   true,
		},
		{
			name:   "typed nil pointer is missing",
			logger: typedNilPtr,
			want:   true,
		},
		{
			name:   "typed nil map is missing",
			logger: typedNilMap,
			want:   true,
		},
		{
			name:   "typed nil func is missing",
			logger: typedNilFunc,
			want:   true,
		},
		{
			name:   "typed nil chan is missing",
			logger: typedNilChan,
			want:   true,
		},
		{
			name:   "typed nil slice is missing",
			logger: typedNilSlice,
			want:   true,
		},
		{
			name:   "real pointer logger is present",
			logger: &testPtrLogger{},
			want:   false,
		},
		{
			name:   "value type logger is present",
			logger: testValueLogger{},
			want:   false,
		},
		{
			name:   "empty non-nil slice logger is present",
			logger: testSliceLogger{},
			want:   false,
		},
		{
			name:   "NoopLogger from constructor is present",
			logger: ports.NewNoopLogger(),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ports.IsNilLogger(tt.logger); got != tt.want {
				t.Errorf("IsNilLogger() = %v, want %v (IsNilLogger must report missing for nil interfaces and typed nils, and present otherwise)", got, tt.want)
			}
		})
	}
}
