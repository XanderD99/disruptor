package handlers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"

	"github.com/XanderD99/discord-disruptor/internal/scheduler"
	"github.com/XanderD99/discord-disruptor/internal/store"
)

type guildReadyTaskBuilder struct {
	store   store.Store
	manager scheduler.Manager
}

func (b *guildReadyTaskBuilder) Build(guildID snowflake.ID, shardID int) guildReadyTask {
	return guildReadyTask{
		guildID: guildID,
		shardID: shardID,
		store:   b.store,
		manager: b.manager,
	}
}

type guildReadyTask struct {
	guildID snowflake.ID
	shardID int
	store   store.Store
	manager scheduler.Manager
}

// Execute implements workerpool.Task.
func (t guildReadyTask) Execute(ctx context.Context) error {
	guild, err := t.store.Guilds().FindByID(ctx, t.guildID.String())
	if err != nil {
		guild, err = t.store.Guilds().Create(ctx, store.Guild{ID: t.guildID.String(), Settings: store.DefaultGuildSettings()})
		if err != nil {
			return fmt.Errorf("failed to create guild %s in store: %w", t.guildID, err)
		}
	}

	// Add guild to voice audio scheduler manager
	if err := t.manager.AddGuild(guild.ID, guild.Settings.Interval); err != nil {
		return fmt.Errorf("failed to add guild %s to voice audio scheduler manager: %w", t.guildID, err)
	}

	return nil
}

func GuildReady(l *slog.Logger, s store.Store, m scheduler.Manager) func(*events.GuildReady) {
	guildReadyTaskBuilder := &guildReadyTaskBuilder{
		store:   s,
		manager: m,
	}

	return func(gr *events.GuildReady) {
		task := guildReadyTaskBuilder.Build(gr.Guild.ID, gr.ShardID())

		go func() {
			if err := task.Execute(context.Background()); err != nil {
				l.Error("Failed to submit guild ready task to worker pool", slog.Any("error", err), slog.String("guild_id", gr.Guild.ID.String()))
				return
			}
		}()
	}
}
