package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type intervalTimer struct {
	timer        *time.Timer
	interval     time.Duration
	nextInterval time.Time
}

func newIntervalTimer(interval time.Duration) *intervalTimer {
	return &intervalTimer{
		timer:        time.NewTimer(interval),
		interval:     interval,
		nextInterval: time.Now().Add(interval),
	}
}

// C returns the timer's channel.
func (t *intervalTimer) C() <-chan time.Time {
	return t.timer.C
}

func (t *intervalTimer) Interval() time.Duration {
	return t.interval
}

// NextInterval returns the next scheduled interval time.
func (t *intervalTimer) NextInterval() time.Time {
	return t.nextInterval
}

// Reset resets the timer and updates the next interval time.
func (t *intervalTimer) Reset() {
	t.nextInterval = time.Now().Add(t.interval)
	t.timer.Reset(t.interval)
}

// Stop stops the timer.
func (t *intervalTimer) Stop() {
	t.timer.Stop()
}

type scheduler struct {
	handler Handler
	logger  *slog.Logger

	mu      sync.RWMutex
	running bool

	stopCh   chan struct{}
	stopOnce sync.Once // Ensures stopCh is closed only once

	timer *intervalTimer
}

func NewScheduler(logger *slog.Logger, handler Handler, opts ...Option[scheduler]) Scheduler {
	interval := time.Hour // Default interval
	for _, opt := range opts {
		interval = opt.IntervalOrDefault(interval)
	}

	g := &scheduler{
		timer:   newIntervalTimer(interval),
		handler: handler,
		stopCh:  make(chan struct{}),
	}

	for _, opt := range opts {
		opt.Apply(g)
	}

	g.logger = logger.With(slog.Any("interval", g.timer.Interval()))

	return g
}

// GetNextIntervalTime returns the next scheduled interval time.
func (ig *scheduler) GetNextIntervalTime() time.Time {
	ig.mu.RLock()
	defer ig.mu.RUnlock()
	if ig.timer == nil {
		return time.Time{}
	}
	return ig.timer.NextInterval()
}

// GetInterval returns the current interval duration.
func (ig *scheduler) GetInterval() time.Duration {
	return ig.timer.Interval()
}

// IsRunning returns true if the scheduler is running.
func (ig *scheduler) IsRunning() bool {
	ig.mu.RLock()
	defer ig.mu.RUnlock()

	return ig.running
}

// Start begins the periodic execution loop.
// Moves timer reset after interval handling and simplifies shutdown logic.
func (ig *scheduler) Start() {
	ig.mu.Lock()
	if ig.running {
		ig.mu.Unlock()
		return
	}
	ig.running = true
	ig.mu.Unlock()

	defer ig.timer.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-ig.stopCh
		cancel()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ig.timer.C():
			start := time.Now()
			err := ig.handler.handle(ctx, ig.timer.Interval())
			duration := time.Since(start)
			ig.logger.Info("Interval handled", slog.Duration("duration", duration))
			if err != nil {
				ig.logger.Error("Failed to handle interval group", slog.Any("error", err))
			}
			ig.timer.Reset() // Reset after handling
		}
	}
}

// Stop gracefully stops the scheduler.
func (ig *scheduler) Stop() error {
	ig.mu.Lock()
	defer ig.mu.Unlock()

	if !ig.running {
		return nil
	}

	ig.running = false
	ig.stopOnce.Do(func() {
		close(ig.stopCh)
	})

	if ig.timer != nil {
		ig.timer.Stop()
		// Drain timer channel to avoid goroutine leaks.
		select {
		case <-ig.timer.C():
		default:
		}
	}
	return nil
}
