package handlers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"

	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/internal/scheduler"
	"github.com/XanderD99/disruptor/pkg/database"
)

type guildReadyTaskBuilder struct {
	store   database.Database
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
	store   database.Database
	manager scheduler.Manager
}

// Execute implements workerpool.Task.
func (t guildReadyTask) Execute(ctx context.Context) error {
	data, err := t.store.FindByID(ctx, t.guildID.String(), &models.Guild{})
	if err != nil {
		data = models.NewGuild(t.guildID)
		if err := t.store.Create(ctx, data); err != nil {
			return fmt.Errorf("failed to create guild %s in store: %w", t.guildID, err)
		}
	}

	guild, ok := data.(*models.Guild)
	if !ok {
		return fmt.Errorf("failed to cast data to models.Guild for guild %s", t.guildID)
	}

	// Add guild to voice audio scheduler manager
	if err := t.manager.AddGuild(guild.ID.String(), guild.Interval); err != nil {
		return fmt.Errorf("failed to add guild %s to voice audio scheduler manager: %w", t.guildID, err)
	}

	return nil
}

func GuildReady(l *slog.Logger, s database.Database, m scheduler.Manager) func(*events.GuildReady) {
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
