package main

import (
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"

	"github.com/XanderD99/discord-disruptor/internal/commands"
	"github.com/XanderD99/discord-disruptor/internal/config"
	"github.com/XanderD99/discord-disruptor/internal/disruptor"
	"github.com/XanderD99/discord-disruptor/internal/handlers"
	"github.com/XanderD99/discord-disruptor/internal/lavalink"
	"github.com/XanderD99/discord-disruptor/internal/metrics"
	"github.com/XanderD99/discord-disruptor/internal/scheduler"
	"github.com/XanderD99/discord-disruptor/internal/store"
	"github.com/XanderD99/discord-disruptor/internal/store/api"
	"github.com/XanderD99/discord-disruptor/pkg/logging"
	"github.com/XanderD99/discord-disruptor/pkg/processes"
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

	pg, store := initStore(cfg)
	pm.AddProcessGroup(pg)

	pg, err = initDiscordProcesses(cfg, logger, store)
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

func initStore(cfg config.Config) (*processes.ProcessGroup, store.Store) {
	group := processes.NewGroup("store", time.Second*5)

	store := api.New(cfg.API)
	group.AddProcessWithCtx("strapi-store", store.Open, false, store.Close)

	return group, store
}

func initDiscordProcesses(cfg config.Config, logger *slog.Logger, store store.Store) (*processes.ProcessGroup, error) {
	group := processes.NewGroup("discord", time.Second*5)

	session, err := disruptor.New(cfg.Token,
		bot.WithShardManagerConfigOpts(
			sharding.WithLogger(logger),
			sharding.WithShardCount(2),
			sharding.WithShardIDs(0, 1),
			sharding.WithAutoScaling(true),
			sharding.WithGatewayConfigOpts(
				gateway.WithIntents(gateway.IntentGuilds, gateway.IntentGuildVoiceStates),
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

	manager := scheduler.NewManager(logger, session, store, lava)
	group.AddProcess("voice-audio-scheduler", manager.Start, false, manager.Stop)

	err = session.AddCommands(
		commands.Play(lava, store),
		commands.Disconnect(lava),
		commands.Next(manager),
		commands.Interval(store, manager),
		commands.Chance(store),
	)
	if err != nil {
		return nil, fmt.Errorf("error adding commands: %w", err)
	}

	session.AddEventListeners(
		bot.NewListenerFunc(handlers.VoiceStateUpdate(logger, lava)),
		bot.NewListenerFunc(handlers.VoiceServerUpdate(logger, lava, store)),

		bot.NewListenerFunc(handlers.GuildJoin(logger, store, manager)),
		bot.NewListenerFunc(handlers.GuildLeave(logger, store, manager)),

		bot.NewListenerFunc(handlers.GuildReady(logger, store, manager)),
	)

	return group, nil
}
