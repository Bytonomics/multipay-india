package logging

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/Bytonomics/multipay-india/multipay-go/ports"
)

// CallerLogger wraps a ports.Logger and automatically captures caller information
// (function name, file name, and line number) for all log messages.
// This ensures that logs always contain context about where they were generated.
type CallerLogger struct {
	delegate ports.Logger
	skip     int // number of stack frames to skip when capturing caller info
}

// NewCallerLogger creates a new CallerLogger that wraps the given logger.
// The skip parameter controls how many stack frames to skip when capturing caller info.
// Typically, skip should be 1 to account for the NewCallerLogger frame itself.
func NewCallerLogger(delegate ports.Logger, skip int) *CallerLogger {
	return &CallerLogger{
		delegate: delegate,
		skip:     skip,
	}
}

// callerInfo returns a formatted string with function name, file name, and line number.
// It uses runtime.Caller to extract this information from the call stack.
func (cl *CallerLogger) callerInfo(depth int) string {
	pc, file, line, ok := runtime.Caller(depth + cl.skip)
	if !ok {
		return "[unknown caller]"
	}

	// Extract function name from the program counter
	funcName := runtime.FuncForPC(pc).Name()
	// Keep only the function name, not the full package path
	if idx := strings.LastIndex(funcName, "."); idx != -1 {
		funcName = funcName[idx+1:]
	}

	// Extract just the filename, not the full path
	if idx := strings.LastIndex(file, "/"); idx != -1 {
		file = file[idx+1:]
	}

	return fmt.Sprintf("[%s:%d %s]", file, line, funcName)
}

// Error logs an error message with caller information automatically prepended.
// Additional key-value pairs can be provided for structured logging context.
func (cl *CallerLogger) Error(ctx context.Context, message string, keysAndValues ...interface{}) {
	caller := cl.callerInfo(1)
	fullMessage := fmt.Sprintf("%s %s", caller, message)
	cl.delegate.Error(ctx, fullMessage, keysAndValues...)
}

// Info logs an info message with caller information automatically prepended.
// Additional key-value pairs can be provided for structured logging context.
func (cl *CallerLogger) Info(ctx context.Context, message string, keysAndValues ...interface{}) {
	caller := cl.callerInfo(1)
	fullMessage := fmt.Sprintf("%s %s", caller, message)
	cl.delegate.Info(ctx, fullMessage, keysAndValues...)
}

// Debug logs a debug message with caller information automatically prepended.
// Additional key-value pairs can be provided for structured logging context.
func (cl *CallerLogger) Debug(ctx context.Context, message string, keysAndValues ...interface{}) {
	caller := cl.callerInfo(1)
	fullMessage := fmt.Sprintf("%s %s", caller, message)
	cl.delegate.Debug(ctx, fullMessage, keysAndValues...)
}

// Verify that CallerLogger implements ports.Logger interface
var _ ports.Logger = (*CallerLogger)(nil)
