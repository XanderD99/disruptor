package scheduler

import (
	"fmt"
	"log/slog"
	"slices"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/disgoorg/snowflake/v2"

	"github.com/XanderD99/disruptor/internal/disruptor"
	"github.com/XanderD99/disruptor/internal/lavalink"
	"github.com/XanderD99/disruptor/pkg/database"
)

func NewManager(logger *slog.Logger, session *disruptor.Session, store database.Database, lavalink lavalink.Lavalink, opts ...Option[manager]) Manager {
	m := &manager{
		intervalGroups:        make(map[string]Scheduler),
		maxGuildsPerScheduler: 100,
		session:               session,
		lavalink:              lavalink,
		store:                 store,
		logger:                logger.With(slog.String("component", "voice_audio_scheduler_manager")),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("starting voice audio scheduler manager", slog.Int("groups", len(m.intervalGroups)))

	for key, group := range m.intervalGroups {
		m.logger.Debug("starting interval group", slog.Group("scheduler", slog.String("key", key), slog.Duration("interval", group.GetInterval())))
		go group.Start()
	}

	return nil
}

func (m *manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("stopping voice audio scheduler manager")

	var eg errgroup.Group
	for key, group := range m.intervalGroups {
		eg.Go(func() error {
			if err := group.Stop(); err != nil {
				m.logger.Error("failed to stop interval group", slog.Group("scheduler", slog.String("key", key)), slog.Any("error", err))
				return fmt.Errorf("failed to stop interval group %s: %w", key, err)
			}
			return nil
		})
	}

	m.intervalGroups = make(map[string]Scheduler)

	if err := eg.Wait(); err != nil {
		m.logger.Error("error stopping some interval groups", slog.Any("error", err))
		return err
	}

	m.logger.Info("voice audio scheduler manager stopped successfully")
	return nil
}

func (m *manager) AddScheduler(opts ...Option[scheduler]) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	handler := NewHandler(m.session, m.store, m.lavalink)
	group := NewScheduler(m.logger, handler, opts...)
	interval := group.GetInterval()

	// Check if we can update an existing scheduler with capacity
	if key, group := m.findSchedulerWithCapacity(interval); group != nil {
		logger := m.logger.With(slog.Group("scheduler", slog.Duration("interval", interval)), slog.String("key", key))

		logger.Debug("updating existing scheduler with capacity")

		if err := group.UpdateOptions(opts...); err != nil {
			m.logger.Error("failed to update interval group options", slog.Any("error", err))
			return fmt.Errorf("failed to update interval group %s: %w", interval, err)
		}

		return nil
	}

	// Generate scheduler key for this interval
	key := m.generateSchedulerKey()
	m.intervalGroups[key] = group
	go group.Start()

	m.logger.Info("new interval group added", slog.Group("scheduler", slog.Duration("interval", interval)), slog.String("key", key))
	return nil
}

func (m *manager) AddGuild(guildID string, interval time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if guild already exists and remove from old scheduler
	if key, group := m.findGuildInSchedulers(guildID); group != nil {
		if group.GetInterval() == interval {
			return nil // Guild already in correct interval
		}

		// Remove from existing group
		if err := group.RemoveGuild(guildID); err != nil {
			m.logger.Error("failed to remove guild from existing interval group",
				slog.String("guild.id", guildID),
				slog.Any("error", err))
			return fmt.Errorf("failed to remove guild from interval group: %w", err)
		}

		// Clean up empty group
		if len(group.GetGuilds()) == 0 {
			if err := group.Stop(); err != nil {
				m.logger.Error("failed to stop empty interval group", slog.Any("error", err))
			}
			delete(m.intervalGroups, key)
		}
	}

	// Find or create a scheduler with capacity for this interval
	schedulerKey, scheduler := m.findOrCreateSchedulerWithCapacity(interval)

	if err := scheduler.AddGuild(guildID); err != nil {
		m.logger.Error("failed to add guild to interval group",
			slog.String("guild.id", guildID),
			slog.Any("error", err))
		return err
	}

	m.logger.Info("guild added to scheduler",
		slog.String("guild.id", guildID),
		slog.Duration("interval", interval),
		slog.String("scheduler_key", schedulerKey))

	return nil
}

func (m *manager) RemoveGuild(guildID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find the scheduler containing this guild
	key, scheduler := m.findGuildInSchedulers(guildID)
	if scheduler == nil {
		return fmt.Errorf("guild %s not found in scheduler", guildID)
	}

	if err := scheduler.RemoveGuild(guildID); err != nil {
		m.logger.Error("failed to remove guild from interval group",
			slog.String("guild.id", guildID),
			slog.Any("error", err))
		return err
	}

	// Clean up empty groups
	if len(scheduler.GetGuilds()) == 0 {
		if err := scheduler.Stop(); err != nil {
			m.logger.Error("failed to stop empty interval group", slog.Any("error", err))
		}
		delete(m.intervalGroups, key)
	}

	m.logger.Info("guild removed from scheduler", slog.String("guild.id", guildID))

	return nil
}

// Public method with locking
func (m *manager) GetSchedulerForGuild(guildID string) (Scheduler, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, scheduler := m.findGuildInSchedulers(guildID)
	if scheduler == nil {
		return nil, fmt.Errorf("guild %s not found in scheduler", guildID)
	}
	return scheduler, nil
}

// findOrCreateSchedulerWithCapacity finds an existing scheduler with capacity or creates a new one
func (m *manager) findOrCreateSchedulerWithCapacity(interval time.Duration) (string, Scheduler) {
	// First, try to find an existing scheduler with capacity for this interval
	if key, scheduler := m.findSchedulerWithCapacity(interval); scheduler != nil {
		return key, scheduler
	}

	handler := NewHandler(m.session, m.store, m.lavalink)
	scheduler := NewScheduler(
		m.logger,
		handler,
		WithInterval(interval),
	)

	// If no existing scheduler has capacity, create a new one
	key := m.generateSchedulerKey()
	m.intervalGroups[key] = scheduler
	go scheduler.Start()

	m.logger.Info("new scheduler created", slog.Group("scheduler",
		slog.String("key", key),
		slog.Duration("interval", interval),
	))

	return key, scheduler
}

// findSchedulerWithCapacity finds a scheduler with available capacity for the given interval
func (m *manager) findSchedulerWithCapacity(interval time.Duration) (string, Scheduler) {
	for key, scheduler := range m.intervalGroups {
		if scheduler.GetInterval() == interval && m.hasCapacity(scheduler) {
			return key, scheduler
		}
	}
	return "", nil
}

// hasCapacity checks if a scheduler has capacity for more guilds
func (m *manager) hasCapacity(scheduler Scheduler) bool {
	if m.maxGuildsPerScheduler <= 0 {
		return true // No limit set
	}

	return len(scheduler.GetGuilds()) < m.maxGuildsPerScheduler
}

// findGuildInSchedulers finds which scheduler contains the given guild
func (m *manager) findGuildInSchedulers(guildID string) (string, Scheduler) {
	for key, scheduler := range m.intervalGroups {
		guilds := scheduler.GetGuilds()
		if slices.Contains(guilds, guildID) {
			return key, scheduler
		}
	}
	return "", nil
}

// generateSchedulerKey creates a unique key for schedulers
func (m *manager) generateSchedulerKey() string {
	return snowflake.New(time.Now()).String()
}
