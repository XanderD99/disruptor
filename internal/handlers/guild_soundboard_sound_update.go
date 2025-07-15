package handlers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/events"

	"github.com/XanderD99/discord-disruptor/internal/models"
	"github.com/XanderD99/discord-disruptor/pkg/database"
)

func GuildSoundBoardSoundUpdate(logger *slog.Logger, db database.Database) func(*events.GuildSoundboardSoundUpdate) {
	logger = logger.With(slog.String("event", "guild_soundboard_update"))

	return func(vsu *events.GuildSoundboardSoundUpdate) {
		logger = logger.With(slog.String("guild.id", vsu.GuildID.String()), slog.String("sound.id", vsu.SoundID.String()))

		ctx := context.Background()

		data, err := db.FindByID(ctx, vsu.SoundID.String(), &models.Sound{})
		if err != nil {
			logger.Warn("Failed to find sound", slog.Any("error", err))
			return
		}

		sound, ok := data.(*models.Sound)
		if !ok {
			logger.Error("Loaded data is not of type *model.Sound", slog.Any("data_type", fmt.Sprintf("%T", data)))
			return
		}

		sound.Name = vsu.Name
		sound.URL = vsu.URL()
		if err := db.Update(ctx, sound); err != nil {
			logger.Error("Failed to update sound", slog.Any("error", err))
			return
		}

		logger.Info("Sound updated successfully")
	}
}
