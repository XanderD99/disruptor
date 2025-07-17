package handlers

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo/events"

	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/internal/scheduler"
	"github.com/XanderD99/disruptor/pkg/database"
)

func GuildLeave(l *slog.Logger, s database.Database, m scheduler.Manager) func(*events.GuildLeave) {
	return func(gr *events.GuildLeave) {
		l = l.With(
			slog.Group("guild", slog.String("id", gr.Guild.ID.String())),
		)

		l.Info("Left guild")

		if err := s.Delete(context.Background(), gr.GuildID.String(), models.Guild{}); err != nil {
			l.Error("Failed to delete guild from store", slog.Any("error", err))
		}

		if err := m.RemoveGuild(gr.Guild.ID.String()); err != nil {
			l.Error("Failed to remove guild from voice audio scheduler manager", slog.Any("error", err))
		}
	}
}
