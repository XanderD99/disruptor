package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bunotel"

	"github.com/XanderD99/disruptor/internal/commands"
	"github.com/XanderD99/disruptor/internal/disruptor"
	"github.com/XanderD99/disruptor/internal/lavalink"
	"github.com/XanderD99/disruptor/internal/listeners"
	"github.com/XanderD99/disruptor/internal/metrics"
	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/internal/otel"
	"github.com/XanderD99/disruptor/internal/scheduler"
	"github.com/XanderD99/disruptor/internal/scheduler/handlers"
	"github.com/XanderD99/disruptor/pkg/logging"
	"github.com/XanderD99/disruptor/pkg/processes"
	"github.com/XanderD99/disruptor/pkg/slogbun"
)

func main() {
	// Note: Context cancellation and graceful shutdown are handled by processes.Manager.
	// This keeps main.go focused on initialization and wiring.

	cfg, err := Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	logger, err := logging.New(cfg.Logging)
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}

	pm := processes.NewManager(logger)

	pg, err := initServerGroup(cfg)
	if err != nil {
		log.Fatalf("Error initializing HTTP servers: %v", err)
	}
	pm.AddProcessGroup(pg)

	// Start system metrics collection
	systemMetrics := metrics.NewSystemMetrics()
	go systemMetrics.StartSystemMetricsCollection(context.Background(), 30*time.Second)

	pg, database, err := initDatabase(cfg, logger)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	pm.AddProcessGroup(pg)

	schedulerGroup, scheduleManager, err := initSchedulers(logger)
	if err != nil {
		log.Fatalf("Error initializing schedulers: %v", err)
	}
	pm.AddProcessGroup(schedulerGroup)

	pg, err = initDiscordProcesses(cfg, logger, database, scheduleManager)
	if err != nil {
		log.Fatalf("Error initializing Discord processes: %v", err)
	}

	pm.AddProcessGroup(pg)

	if err := pm.Run(); err != nil {
		log.Fatalf("Error running process manager: %v", err)
	}
}

func initServerGroup(cfg Config) (*processes.ProcessGroup, error) {
	group := processes.NewGroup("servers", time.Second*5)

	// Initialize OpenTelemetry metrics and get the Prometheus exporter
	promExporter, err := otel.InitMetrics(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error initializing OpenTelemetry metrics: %w", err)
	}

	// Initialize metrics server with OpenTelemetry Prometheus exporter
	metricsServer, err := metrics.NewServerWithExporter(cfg.Metrics, promExporter)
	if err != nil {
		return nil, fmt.Errorf("error creating metrics server: %w", err)
	}
	group.AddProcessWithCtx("metrics", metricsServer.Run, false, nil)

	return group, nil
}

func initSchedulers(logger *slog.Logger) (*processes.ProcessGroup, *scheduler.Manager, error) {
	group := processes.NewGroup("schedulers", time.Second*5)

	// Initialize voice audio scheduler
	voiceAudioScheduler := scheduler.NewManager(scheduler.WithLogger(logger))

	group.AddProcessWithCtx("manager", voiceAudioScheduler.Start, false, voiceAudioScheduler.Stop)

	return group, voiceAudioScheduler, nil
}

func initDatabase(cfg Config, logger *slog.Logger) (*processes.ProcessGroup, *bun.DB, error) {
	group := processes.NewGroup("database", time.Second*5)

	// Initialize database connection
	var database *bun.DB

	switch cfg.Database.Type {
	case "sqlite":
		sqldb, err := sql.Open(sqliteshim.ShimName, cfg.Database.DSN)
		if err != nil {
			return nil, nil, fmt.Errorf("error initializing database: %w", err)
		}
		group.AddProcessWithoutStart("sqlite", sqldb.Close)

		database = bun.NewDB(sqldb, sqlitedialect.New())
	default:
		return group, nil, fmt.Errorf("invalid database type: %s", cfg.Database.Type)
	}

	// Add OpenTelemetry hook for automatic metrics collection
	database.AddQueryHook(bunotel.NewQueryHook(
		bunotel.WithDBName("disruptor"),
		bunotel.WithAttributes(
			// Add additional attributes if needed
		),
	))

	// Add slog hook for logging (without custom metrics)
	database.AddQueryHook(slogbun.NewQueryHook(
		slogbun.WithLogger(logger),
	))

	group.AddProcessWithCtx("database", func(ctx context.Context) error {
		_, err := database.NewCreateTable().IfNotExists().Model(&models.Guild{}).Exec(ctx)
		return err
	}, false, nil)

	return group, database, nil
}

func initDiscordProcesses(cfg Config, logger *slog.Logger, db *bun.DB, scheduleManager *scheduler.Manager) (*processes.ProcessGroup, error) {
	group := processes.NewGroup("discord", time.Second*5)

	session, err := disruptor.New(cfg.Disruptor)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord bot: %w", err)
	}
	group.AddProcessWithCtx("session", session.Open, false, session.Close)

	lava := lavalink.New(cfg.LavalinkNodes, session, logger)
	group.AddProcessWithCtx("disgolink", lava.Start, false, nil)

	scheduleManager.RegisterBuilder(handlers.HandlerTypeRandomVoiceJoin, func(interval time.Duration) *scheduler.Scheduler {
		return scheduler.NewScheduler(interval, handlers.NewRandomVoiceJoinHandler(session, db))
	})

	err = session.AddCommands(
		commands.Play(lava),
		commands.Disconnect(lava),
		commands.Next(db, scheduleManager),
		commands.Interval(db, scheduleManager),
		commands.Chance(db),
	)
	if err != nil {
		return nil, fmt.Errorf("error adding commands: %w", err)
	}

	session.AddEventListeners(
		bot.NewListenerFunc(listeners.VoiceStateUpdate(logger, lava)),
		bot.NewListenerFunc(listeners.VoiceServerUpdate(logger, lava)),
		bot.NewListenerFunc(listeners.GuildJoin(logger, db, scheduleManager)),
		bot.NewListenerFunc(listeners.GuildLeave(logger, db, scheduleManager)),
		bot.NewListenerFunc(listeners.GuildReady(logger, db, scheduleManager)),
	)

	// Initialize Discord collector for guild metrics
	discordCollector := metrics.NewDiscordCollector(session)
	group.AddProcessWithCtx("discord_metrics", discordCollector.StartCollection, false, nil)

	return group, nil
}
