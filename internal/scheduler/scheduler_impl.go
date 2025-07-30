package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type scheduler struct {
	interval time.Duration

	handler Handler
	logger  *slog.Logger

	mu               sync.RWMutex
	running          bool
	stopCh           chan struct{}
	timer            *time.Timer
	nextIntervalTime time.Time
}

func NewScheduler(logger *slog.Logger, handler Handler, opts ...Option[scheduler]) Scheduler {
	g := &scheduler{
		interval: time.Hour,
		handler:  handler,
		logger:   logger.With(slog.Group("scheduler", slog.Duration("interval", time.Hour))),

		stopCh: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

func (ig *scheduler) UpdateOptions(opts ...Option[scheduler]) error {
	ig.mu.Lock()
	defer ig.mu.Unlock()

	for _, opt := range opts {
		opt(ig)
	}

	return nil
}

func (ig *scheduler) GetNextIntervalTime() time.Time {
	ig.mu.RLock()
	defer ig.mu.RUnlock()
	if ig.timer == nil {
		return time.Time{}
	}
	return ig.nextIntervalTime
}

func (ig *scheduler) Start() {
	// Only lock for initialization
	ig.mu.Lock()
	if ig.running {
		ig.mu.Unlock()
		return
	}
	ig.running = true
	ig.mu.Unlock() // ❌ Release lock before entering loop!

	ig.timer = time.NewTimer(ig.interval)
	defer ig.timer.Stop()

	for {
		ig.nextIntervalTime = time.Now().Add(ig.interval)
		select {
		case <-ig.stopCh:
			return
		case <-ig.timer.C:
			if err := ig.handler.handle(context.Background(), ig.interval); err != nil {
				ig.logger.Error("Failed to handle interval group", slog.Any("error", err))
			}

			ig.timer.Reset(ig.interval)
		}
	}
}

func (ig *scheduler) GetInterval() time.Duration {
	return ig.interval
}

func (ig *scheduler) Stop() error {
	ig.mu.Lock()
	defer ig.mu.Unlock()

	if !ig.running {
		return nil
	}

	ig.running = false
	close(ig.stopCh)

	if ig.timer != nil {
		ig.timer.Stop()
	}
	return nil
}

func (ig *scheduler) IsRunning() bool {
	ig.mu.RLock()
	defer ig.mu.RUnlock()

	return ig.running
}
