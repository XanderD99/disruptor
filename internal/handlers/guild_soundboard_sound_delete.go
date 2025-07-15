package handlers

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo/events"

	"github.com/XanderD99/discord-disruptor/internal/models"
	"github.com/XanderD99/discord-disruptor/pkg/database"
)

func GuildSoundBoardSoundDelete(logger *slog.Logger, db database.Database) func(*events.GuildSoundboardSoundDelete) {
	logger = logger.With(slog.String("event", "guild_soundboard_delete"))

	return func(vsd *events.GuildSoundboardSoundDelete) {
		logger = logger.With(slog.String("guild.id", vsd.GuildID.String()), slog.String("sound.id", vsd.SoundID.String()))

		ctx := context.Background()

		if err := db.Delete(ctx, vsd.SoundID.String(), &models.Sound{}); err != nil {
			logger.Error("Failed to delete sound", slog.Any("error", err))
			return
		}

		logger.Info("Sound deleted successfully")
	}
}
