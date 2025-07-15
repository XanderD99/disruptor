package handlers

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo/events"

	"github.com/XanderD99/discord-disruptor/internal/scheduler"
	"github.com/XanderD99/discord-disruptor/internal/store"
)

func GuildLeave(l *slog.Logger, s store.Store, m scheduler.Manager) func(*events.GuildLeave) {
	return func(gr *events.GuildLeave) {
		l = l.With(
			slog.Group("guild", slog.String("id", gr.Guild.ID.String())),
		)

		l.Info("Guild removed")

		ctx := context.Background()
		if err := s.Guilds().Delete(ctx, gr.Guild.ID.String()); err != nil {
			l.Error("Failed to delete guild from store", slog.Any("error", err))
		}

		if err := m.RemoveGuild(gr.Guild.ID.String()); err != nil {
			l.Error("Failed to remove guild from voice audio scheduler manager", slog.Any("error", err))
		}
	}
}
