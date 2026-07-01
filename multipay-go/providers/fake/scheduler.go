package fake

import (
	"sync"
	"time"
)

// Scheduler decouples "run fn after delay" so the delay mechanism (wall-clock,
// immediate, or a custom impl) is injectable into the Harness.
type Scheduler interface {
	// Schedule arranges for fn to run after delay. It returns immediately.
	Schedule(delay time.Duration, fn func())
	// Wait blocks until every function scheduled so far has finished running.
	Wait()
}

// compile-time interface checks
var (
	_ Scheduler = (*WallClockScheduler)(nil)
	_ Scheduler = (*ImmediateScheduler)(nil)
)

// WallClockScheduler runs each fn after a REAL time.Duration using time.AfterFunc.
// Wait blocks until all scheduled fns have completed. It is the default Scheduler.
type WallClockScheduler struct {
	wg sync.WaitGroup
}

// NewWallClockScheduler returns a wall-clock scheduler.
func NewWallClockScheduler() *WallClockScheduler {
	return &WallClockScheduler{}
}

// Schedule runs fn after delay on a timer goroutine.
func (s *WallClockScheduler) Schedule(delay time.Duration, fn func()) {
	s.wg.Add(1)
	time.AfterFunc(delay, func() {
		defer s.wg.Done()
		fn()
	})
}

// Wait blocks until all scheduled fns have finished.
func (s *WallClockScheduler) Wait() {
	s.wg.Wait()
}

// ImmediateScheduler runs fn synchronously inside Schedule, ignoring delay.
// Useful for fast, fully-synchronous tests.
type ImmediateScheduler struct{}

// NewImmediateScheduler returns a scheduler that runs work synchronously.
func NewImmediateScheduler() *ImmediateScheduler {
	return &ImmediateScheduler{}
}

// Schedule runs fn synchronously, ignoring delay.
func (s *ImmediateScheduler) Schedule(delay time.Duration, fn func()) {
	fn()
}

// Wait is a no-op for the immediate scheduler.
func (s *ImmediateScheduler) Wait() {}
