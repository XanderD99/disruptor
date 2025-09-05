package listeners

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/uptrace/bun"

	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/internal/scheduler"
	"github.com/XanderD99/disruptor/internal/scheduler/handlers"
)

type guildReadyTaskBuilder struct {
	db      *bun.DB
	manager *scheduler.Manager
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
	db      *bun.DB
	manager *scheduler.Manager
}

// Execute implements workerpool.Task.
func (t guildReadyTask) Execute(ctx context.Context) error {
	guild := models.NewGuild(t.guildID)

	if err := t.db.NewInsert().Model(&guild).Ignore().Scan(ctx); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to insert guild %s into database: %w", t.guildID, err)
	}

	if err := t.db.NewSelect().Model(&guild).WherePK().Scan(ctx); err != nil {
		return fmt.Errorf("failed to fetch guild %s from database: %w", t.guildID, err)
	}

	if err := t.manager.AddScheduler(handlers.HandlerTypeRandomVoiceJoin, guild.Interval); err != nil {
		return fmt.Errorf("failed to add guild %s to voice audio scheduler manager: %w", t.guildID, err)
	}

	return nil
}

func GuildReady(l *slog.Logger, db *bun.DB, m *scheduler.Manager) func(*events.GuildReady) {
	guildReadyTaskBuilder := &guildReadyTaskBuilder{
		db:      db,
		manager: m,
	}

	return func(gr *events.GuildReady) {
		task := guildReadyTaskBuilder.Build(gr.Guild.ID, gr.ShardID())

		go func() {
			if err := task.Execute(context.Background()); err != nil {
				l.Error("Failed to submit guild ready task to worker pool", slog.Any("error", err), slog.String("guild.id", gr.Guild.ID.String()))
				return
			}
		}()
	}
}
