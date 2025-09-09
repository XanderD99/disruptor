package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/XanderD99/disruptor/internal/util"
	"github.com/XanderD99/disruptor/pkg/logging"
)

type HandleFunc func(ctx context.Context) error

type Scheduler struct {
	handler HandleFunc

	mu      sync.RWMutex
	running bool

	stopCh   chan struct{}
	stopOnce sync.Once // Ensures stopCh is closed only once

	timer        *time.Timer
	interval     time.Duration
	nextInterval time.Time
}

func NewScheduler(interval time.Duration, handler HandleFunc) *Scheduler {
	g := &Scheduler{
		timer:        time.NewTimer(interval),
		interval:     interval,
		nextInterval: time.Now().Add(interval),
		handler:      handler,
		stopCh:       make(chan struct{}),
	}

	return g
}

func (t *Scheduler) Interval() time.Duration {
	return t.interval
}

// NextInterval returns the next scheduled interval time.
func (t *Scheduler) NextInterval() time.Time {
	return t.nextInterval
}

// Reset resets the timer and updates the next interval time.
func (t *Scheduler) Reset() {
	t.nextInterval = time.Now().Add(t.interval)
	t.timer.Reset(t.interval)
}

// GetInterval returns the current interval duration.
func (ig *Scheduler) GetInterval() time.Duration {
	return ig.Interval()
}

// IsRunning returns true if the scheduler is running.
func (ig *Scheduler) IsRunning() bool {
	ig.mu.RLock()
	defer ig.mu.RUnlock()

	return ig.running
}

// Start begins the periodic execution loop.
// Moves timer reset after interval handling and simplifies shutdown logic.
func (ig *Scheduler) Start(ctx context.Context) {
	ig.mu.Lock()
	if ig.running {
		ig.mu.Unlock()
		return
	}
	ig.running = true
	ig.mu.Unlock()

	defer ig.timer.Stop()

	go func() {
		<-ig.stopCh
	}()

	logger := logging.FromContext(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ig.timer.C:
			// add interval to context
			ctx = util.AddIntervalToContext(ctx, ig.Interval())

			if err := ig.handler(ctx); err != nil {
				logger.ErrorContext(ctx, "Failed to handle interval group", slog.Any("error", err))
			}

			ig.Reset() // Reset after handling
		}
	}
}

// Stop gracefully stops the scheduler.
func (ig *Scheduler) Stop() error {
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
		case <-ig.timer.C:
		default:
		}
	}
	return nil
}
