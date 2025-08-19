package scheduler

import (
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/uptrace/bun"

	"github.com/XanderD99/disruptor/internal/disruptor"
	"github.com/XanderD99/disruptor/internal/lavalink"
)

func NewManager(logger *slog.Logger, session *disruptor.Session, db *bun.DB, lavalink lavalink.Lavalink) Manager {
	m := &manager{
		schedulers: make(map[time.Duration]Scheduler),
		session:    session,
		lavalink:   lavalink,
		db:         db,
		logger:     logger.With(slog.String("component", "voice_audio_scheduler_manager")),
	}

	return m
}

func (m *manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("starting voice audio scheduler manager", slog.Int("groups", len(m.schedulers)))

	for interval, group := range m.schedulers {
		m.logger.Debug("starting interval group", slog.Group("scheduler", slog.Duration("interval", interval)))
		go group.Start()
	}

	return nil
}

func (m *manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("stopping voice audio scheduler manager")

	var eg errgroup.Group
	for interval, group := range m.schedulers {
		eg.Go(func() error {
			if err := group.Stop(); err != nil {
				m.logger.Error("failed to stop interval group", slog.Group("scheduler", slog.Duration("interval", interval)), slog.Any("error", err))
				return fmt.Errorf("failed to stop interval group %v: %w", interval, err)
			}
			return nil
		})
	}

	m.schedulers = make(map[time.Duration]Scheduler)

	if err := eg.Wait(); err != nil {
		m.logger.Error("error stopping some interval groups", slog.Any("error", err))
		return err
	}

	m.logger.Info("voice audio scheduler manager stopped successfully")
	return nil
}

func (m *manager) GetScheduler(interval time.Duration) (Scheduler, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if scheduler, exists := m.schedulers[interval]; exists {
		return scheduler, true
	}
	return nil, false
}

func (m *manager) AddScheduler(opts ...Option[scheduler]) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	handler := NewHandler(m.session, m.db, m.lavalink)
	group := NewScheduler(m.logger, handler, opts...)
	interval := group.GetInterval()

	// Check if scheduler for this interval already exists
	if _, exists := m.schedulers[interval]; exists {
		m.logger.Debug("updating existing scheduler", slog.Duration("interval", interval))
		return nil
	}

	// Create new scheduler for this interval
	m.schedulers[interval] = group
	go group.Start()

	m.logger.Info("new interval group added", slog.Duration("interval", interval))
	return nil
}
