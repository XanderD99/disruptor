package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/XanderD99/disruptor/internal/metrics"
	"github.com/XanderD99/disruptor/pkg/logging"
)

func NewManager(opts ...Option[Manager]) *Manager {
	m := &Manager{
		schedulers:       make(map[string]*Scheduler),
		builders:         make(map[string]SchedulerBuilder),
		logger:           slog.Default(),
		schedulerMetrics: metrics.NewSchedulerMetrics(),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

type Option[T any] func(*T)

func WithLogger(logger *slog.Logger) Option[Manager] {
	return func(m *Manager) {
		m.logger = logger
	}
}

func WithBuilder(key string, builder SchedulerBuilder) Option[Manager] {
	return func(m *Manager) {
		m.builders[key] = builder
	}
}

type SchedulerBuilder func(interval time.Duration) *Scheduler

type Manager struct {
	schedulers map[string]*Scheduler
	builders   map[string]SchedulerBuilder

	// Dependencies
	logger           *slog.Logger
	schedulerMetrics *metrics.SchedulerMetrics

	mu sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
}

func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancel != nil {
		m.cancel() // Cancel any previous context
	}

	ctx = logging.AddToContext(ctx, m.logger)
	m.ctx, m.cancel = context.WithCancel(ctx)

	m.logger.InfoContext(m.ctx, "starting voice audio scheduler manager", slog.Int("groups", len(m.schedulers)))

	for key, group := range m.schedulers {
		m.logger.DebugContext(m.ctx, "starting interval group", slog.Group("scheduler", slog.String("key", key)))
		go group.Start(m.ctx)
	}

	return nil
}

func (m *Manager) RegisterBuilder(key string, builder SchedulerBuilder) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.builders == nil {
		m.builders = make(map[string]SchedulerBuilder)
	}
	m.builders[key] = builder
}

// Graceful shutdown: stop all schedulers, wait for them to finish, then clear the map.
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("stopping voice audio scheduler manager")

	if m.cancel != nil {
		m.cancel() // Cancel context for all groups
	}

	var eg errgroup.Group
	schedulersCopy := make(map[string]*Scheduler, len(m.schedulers))
	maps.Copy(schedulersCopy, m.schedulers)

	for key, group := range schedulersCopy {
		eg.Go(func() error {
			if err := group.Stop(); err != nil {
				m.logger.Error("failed to stop interval group", slog.Group("scheduler", slog.String("key", key)), slog.Any("error", err))
				return fmt.Errorf("failed to stop interval group %v: %w", key, err)
			}
			return nil
		})
	}

	// Wait for all schedulers to stop before clearing the map
	if err := eg.Wait(); err != nil {
		m.logger.Error("error stopping some interval groups", slog.Any("error", err))
		return err
	}

	m.schedulers = make(map[string]*Scheduler)

	// Update queue depth metrics to 0
	m.schedulerMetrics.UpdateQueueDepth(0)

	m.logger.Info("voice audio scheduler manager stopped successfully")
	return nil
}

func (m *Manager) GetScheduler(scheduler string, interval time.Duration) (*Scheduler, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := fmt.Sprintf("%s_%d", scheduler, interval.Milliseconds())
	if scheduler, exists := m.schedulers[key]; exists {
		return scheduler, true
	}
	return nil, false
}

func (m *Manager) AddScheduler(key string, interval time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	schedKey := fmt.Sprintf("%s_%d", key, interval.Milliseconds())
	if _, exists := m.schedulers[schedKey]; exists {
		m.logger.Warn("scheduler with this interval already exists", slog.String("key", key), slog.Duration("interval", interval))
		return nil
	}

	builder, ok := m.builders[key]
	if !ok {
		m.logger.Error("no builder registered for key", slog.String("key", key))
		return fmt.Errorf("no builder registered for key %s", key)
	}

	group := builder(interval)
	m.schedulers[schedKey] = group

	// Update queue depth metrics
	m.schedulerMetrics.UpdateQueueDepth(int64(len(m.schedulers)))

	m.logger.Info("new interval group added", slog.String("key", key), slog.Duration("interval", interval))
	if m.ctx != nil {
		go group.Start(m.ctx)
	}

	return nil
}
