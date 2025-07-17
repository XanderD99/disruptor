package handlers

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo/events"

	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/internal/scheduler"
	"github.com/XanderD99/disruptor/pkg/database"
)

func GuildJoin(l *slog.Logger, s database.Database, m scheduler.Manager) func(*events.GuildJoin) {

	return func(gj *events.GuildJoin) {
		l = l.With(slog.Group("guild", slog.String("id", gj.Guild.ID.String())))

		l.Info("Joined guild")

		guild := models.NewGuild(gj.Guild.ID)

		if err := s.Create(context.Background(), &guild); err != nil {
			l.Error("Failed to create guild in store", slog.Any("error", err))
			return
		}

		if err := m.AddGuild(guild.ID.String(), guild.Interval); err != nil {
			l.Error("Failed to add guild to voice audio scheduler manager", slog.Any("error", err))
		}
	}
}
