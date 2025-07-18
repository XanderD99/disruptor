package handlers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"

	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/internal/scheduler"
	"github.com/XanderD99/disruptor/pkg/db"
)

type guildReadyTaskBuilder struct {
	db      db.Database
	manager scheduler.Manager
}

func (b *guildReadyTaskBuilder) Build(guildID snowflake.ID, shardID int) guildReadyTask {
	return guildReadyTask{
		guildID: guildID,
		shardID: shardID,
		db:      b.db,
		manager: b.manager,
	}
}

type guildReadyTask struct {
	guildID snowflake.ID
	shardID int
	db      db.Database
	manager scheduler.Manager
}

// Execute implements workerpool.Task.
func (t guildReadyTask) Execute(ctx context.Context) error {
	var guild models.Guild

	if err := t.db.FindByID(ctx, t.guildID, &guild); err != nil {
		guild = models.NewGuild(t.guildID)
		if err := t.db.Create(ctx, guild); err != nil {
			return fmt.Errorf("failed to create guild %s in store: %w", t.guildID, err)
		}
	}

	// Add guild to voice audio scheduler manager
	if err := t.manager.AddGuild(guild.ID.String(), guild.Interval); err != nil {
		return fmt.Errorf("failed to add guild %s to voice audio scheduler manager: %w", t.guildID, err)
	}

	return nil
}

func GuildReady(l *slog.Logger, db db.Database, m scheduler.Manager) func(*events.GuildReady) {
	guildReadyTaskBuilder := &guildReadyTaskBuilder{
		db:      db,
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
