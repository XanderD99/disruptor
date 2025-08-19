package listeners

import (
	"context"
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/events"
	"github.com/uptrace/bun"
	"golang.org/x/sync/errgroup"

	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/internal/scheduler"
)

func GuildJoin(l *slog.Logger, db *bun.DB, m scheduler.Manager) func(*events.GuildJoin) {
	return func(gj *events.GuildJoin) {
		l = l.With(slog.Group("guild", slog.String("id", gj.Guild.ID.String())))

		l.Info("Joined guild")

		guild := models.NewGuild(gj.Guild.ID)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var errGroup errgroup.Group

		errGroup.Go(func() (err error) {
			_, err = db.NewInsert().Model(&guild).Exec(ctx)
			return
		})

		errGroup.Go(func() (err error) {
			err = m.AddScheduler(scheduler.WithInterval(guild.Interval))
			return
		})

		if err := errGroup.Wait(); err != nil {
			l.Error("Failed to join guild", slog.Any("error", err))
		}
	}
}
