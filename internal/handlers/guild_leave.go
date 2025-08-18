package handlers

import (
	"context"
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/events"
	"github.com/uptrace/bun"

	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/internal/scheduler"
)

func GuildLeave(l *slog.Logger, db *bun.DB, m scheduler.Manager) func(*events.GuildLeave) {
	return func(gr *events.GuildLeave) {
		l = l.With(slog.Group("guild", slog.String("id", gr.Guild.ID.String())))

		l.Info("Left guild")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if _, err := db.NewDelete().Model(&models.Guild{Snowflake: gr.Guild.ID}).Exec(ctx); err != nil {
			l.Error("Failed to remove guild from store", slog.Any("error", err))
		}
	}
}
