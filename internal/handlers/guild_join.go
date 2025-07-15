package handlers

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo/events"

	"github.com/XanderD99/discord-disruptor/internal/scheduler"
	"github.com/XanderD99/discord-disruptor/internal/store"
)

func GuildJoin(l *slog.Logger, s store.Store, m scheduler.Manager) func(*events.GuildJoin) {
	settings := store.DefaultGuildSettings()

	return func(gj *events.GuildJoin) {
		l = l.With(
			slog.Group("guild", slog.String("id", gj.Guild.ID.String())),
		)

		ctx := context.Background()
		guild, err := s.Guilds().Create(ctx, store.Guild{ID: gj.Guild.ID.String(), Settings: settings})
		if err != nil {
			l.Error("Failed to create guild in store", slog.Any("error", err))
		}

		if err := m.AddGuild(guild.ID, settings.Interval); err != nil {
			l.Error("Failed to add guild to voice audio scheduler manager", slog.Any("error", err))
		}
	}
}
