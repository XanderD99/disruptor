package handlers

import (
	"context"
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/events"

	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/internal/scheduler"
	"github.com/XanderD99/disruptor/pkg/db"
)

func GuildLeave(l *slog.Logger, d db.Database, m scheduler.Manager) func(*events.GuildLeave) {
	return func(gr *events.GuildLeave) {
		l = l.With(
			slog.Group("guild", slog.String("id", gr.Guild.ID.String())),
		)

		l.Info("Left guild")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := d.FindOne(ctx, &models.Guild{ID: gr.Guild.ID}); err != nil {
			l.Error("Failed to find guild in store", slog.Any("error", err))
			return
		}
	}
}
