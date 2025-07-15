package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/events"
	disgolavalink "github.com/disgoorg/disgolink/v3/lavalink"

	"github.com/XanderD99/discord-disruptor/internal/lavalink"
	"github.com/XanderD99/discord-disruptor/internal/models"
	"github.com/XanderD99/discord-disruptor/pkg/database"
	"github.com/XanderD99/discord-disruptor/pkg/util"
)

func VoiceServerUpdate(logger *slog.Logger, lava lavalink.Lavalink, db database.Database) func(*events.VoiceServerUpdate) {
	logger = logger.With(slog.String("event", "voice_server_update"))

	return func(vsu *events.VoiceServerUpdate) {
		logger = logger.With(slog.String("guild.id", vsu.GuildID.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		lava.OnVoiceServerUpdate(ctx, vsu.GuildID, vsu.Token, *vsu.Endpoint)

		filter := map[string]any{
			"$or": []any{
				map[string]any{"guild_id": vsu.GuildID.String()},
				map[string]any{"guild_id": nil},
			},
		}

		data, err := db.FindAll(ctx, models.Sound{}, database.WithFilters(filter))
		if err != nil {
			logger.Error("Failed to find soundboard sounds", slog.Any("error", err))
			return
		}

		sounds, ok := data.([]models.Sound)
		if !ok {
			logger.Error("Loaded data is not of type []model.Sound", slog.Any("data_type", fmt.Sprintf("%T", data)))
			return
		}

		if len(sounds) == 0 {
			logger.Info("No soundboard sounds found for guild, skipping track update")
			return
		}

		// pick random sound from the list
		index := util.RandomInt(0, len(sounds)-1)
		sound := sounds[index]

		results, err := lava.BestNode().LoadTracks(ctx, sound.URL)
		if err != nil {
			logger.Error("Failed to load tracks", slog.Any("error", err), slog.String("sound_url", sound.URL))
			return
		}

		track, ok := results.Data.(disgolavalink.Track)
		if !ok {
			logger.Error("Loaded track is not of type disgolavalink.Track", slog.Any("data_type", fmt.Sprintf("%T", results.Data)))
			return
		}

		player := lava.Player(vsu.GuildID)
		if err := player.Update(ctx, disgolavalink.WithTrack(track)); err != nil {
			logger.Error("Failed to update player with new track", slog.Any("error", err))
			return
		}
	}
}
