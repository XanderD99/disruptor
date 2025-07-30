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

func GuildJoin(l *slog.Logger, d db.Database, m scheduler.Manager) func(*events.GuildJoin) {
	return func(gj *events.GuildJoin) {
		l = l.With(slog.Group("guild", slog.String("id", gj.Guild.ID.String())))

		l.Info("Joined guild")

		guild := models.NewGuild(gj.Guild.ID)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := d.Create(ctx, guild); err != nil {
			l.Error("Failed to create guild in store", slog.Any("error", err))
			return
		}

		if err := m.AddScheduler(scheduler.WithInterval(guild.Interval)); err != nil {
			l.Error("Failed to add guild to voice audio scheduler manager", slog.Any("error", err))
		}
	}
}
