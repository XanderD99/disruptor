package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"

	"github.com/XanderD99/disruptor/internal/commands"
	"github.com/XanderD99/disruptor/internal/config"
	"github.com/XanderD99/disruptor/internal/disruptor"
	"github.com/XanderD99/disruptor/internal/lavalink"
	"github.com/XanderD99/disruptor/internal/listeners"
	"github.com/XanderD99/disruptor/internal/metrics"
	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/internal/scheduler"
	"github.com/XanderD99/disruptor/pkg/logging"
	"github.com/XanderD99/disruptor/pkg/processes"
	"github.com/XanderD99/disruptor/pkg/slogbun"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	logger, err := logging.New(cfg.Logging)
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}

	pm := processes.NewManager(logger)

	pg, err := httpServers(cfg)
	if err != nil {
		log.Fatalf("Error initializing HTTP servers: %v", err)
	}
	pm.AddProcessGroup(pg)

	pg, database, err := initDatabase(cfg, logger)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	pm.AddProcessGroup(pg)

	pg, err = initDiscordProcesses(cfg, logger, database)
	if err != nil {
		log.Fatalf("Error initializing Discord processes: %v", err)
	}

	pm.AddProcessGroup(pg)

	if err := pm.Run(); err != nil {
		log.Fatalf("Error running process manager: %v", err)
	}
}

func httpServers(cfg config.Config) (*processes.ProcessGroup, error) {
	group := processes.NewGroup("http", time.Second*5)

	// Initialize metrics server
	metricsServer, err := metrics.NewServer(cfg.Metrics)
	if err != nil {
		return nil, fmt.Errorf("error creating metrics server: %w", err)
	}
	group.AddProcessWithCtx("metrics-server", metricsServer.Run, false, nil)

	return group, nil
}

func initDatabase(cfg config.Config, logger *slog.Logger) (*processes.ProcessGroup, *bun.DB, error) {
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

	database.AddQueryHook(slogbun.NewQueryHook(
		slogbun.WithLogger(logger),
	))

	group.AddProcessWithCtx("database", func(ctx context.Context) error {
		_, err := database.NewCreateTable().IfNotExists().Model(&models.Guild{}).Exec(ctx)
		return err
	}, false, nil)

	return group, database, nil
}

func initDiscordProcesses(cfg config.Config, logger *slog.Logger, db *bun.DB) (*processes.ProcessGroup, error) {
	group := processes.NewGroup("discord", time.Second*5)

	session, err := disruptor.New(cfg.Token,
		bot.WithShardManagerConfigOpts(
			sharding.WithLogger(logger),
			sharding.WithShardCount(2),
			sharding.WithShardIDs(0, 1),
			sharding.WithAutoScaling(true),
			sharding.WithGatewayConfigOpts(
				gateway.WithIntents(gateway.IntentGuilds, gateway.IntentGuildVoiceStates, gateway.IntentGuildExpressions),
				gateway.WithCompress(true),
				gateway.WithPresenceOpts(
					gateway.WithListeningActivity("to your commands"),
				),
			),
		),
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagsAll),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord bot: %w", err)
	}
	group.AddProcessWithCtx("bot", session.Open, false, session.Close)

	lava := lavalink.New(cfg.LavalinkNodes, session, logger)
	group.AddProcessWithCtx("disgolink", lava.Start, false, nil)

	manager := scheduler.NewManager(logger, session, db, lava)
	group.AddProcess("voice-audio-scheduler", manager.Start, false, manager.Stop)

	err = session.AddCommands(
		commands.Play(lava),
		commands.Disconnect(lava),
		commands.Next(manager, db),
		commands.Interval(db, manager),
		commands.Chance(db),
	)
	if err != nil {
		return nil, fmt.Errorf("error adding commands: %w", err)
	}

	session.AddEventListeners(
		bot.NewListenerFunc(listeners.VoiceStateUpdate(logger, lava)),
		bot.NewListenerFunc(listeners.VoiceServerUpdate(logger, lava)),
		bot.NewListenerFunc(listeners.GuildJoin(logger, db, manager)),
		bot.NewListenerFunc(listeners.GuildLeave(logger, db, manager)),
		bot.NewListenerFunc(listeners.GuildReady(logger, db, manager)),
	)

	return group, nil
}
