package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/uptrace/bun"

	"github.com/XanderD99/disruptor/internal/disruptor"
	"github.com/XanderD99/disruptor/internal/metrics"
	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/internal/scheduler"
	"github.com/XanderD99/disruptor/internal/util"
)

const HandlerTypeRandomVoiceJoin = "random_voice_join"

func NewRandomVoiceJoinHandler(session *disruptor.Session, db *bun.DB) scheduler.HandleFunc {
	registerHandlerSingleton(HandlerTypeRandomVoiceJoin, func() any {
		return newRandomVoiceJoinHandler(session, db)
	})

	cb, ok := getHandlerSingleton(HandlerTypeRandomVoiceJoin).(scheduler.HandleFunc)
	if !ok {
		panic(fmt.Sprintf("handler %s is not a scheduler.HandleFunc", HandlerTypeRandomVoiceJoin))
	}
	return cb
}

func newRandomVoiceJoinHandler(session *disruptor.Session, db *bun.DB) scheduler.HandleFunc {
	// Import the metrics package at the top if not already imported
	originalHandler := func(ctx context.Context) error {
		chance := util.RandomFloat(0, 100) // Use float for better precision

		interval, ok := util.GetIntervalFromContext(ctx)
		if !ok {
			return fmt.Errorf("failed to get interval from context")
		}

		// Create a context with timeout for this batch
		guilds, err := getEligibleGuilds(ctx, db, interval, chance)
		if err != nil {
			return fmt.Errorf("failed to find guilds: %w", err)
		}

		maxWorkers := int(math.Max(1, math.Sqrt(float64(len(guilds)))))

		// Process guilds with worker pool
		return util.ProcessWithWorkerPool(ctx, guilds, maxWorkers, func(ctx context.Context, guild models.Guild) {
			if err := processGuild(ctx, session, guild.Snowflake); err != nil {
				session.Logger().ErrorContext(ctx, "Failed to process guild", slog.Any("guild.id", guild.Snowflake), slog.Any("error", err))
			}
		})
	}

	// Wrap with metrics collection
	return metrics.WithJobMetrics(HandlerTypeRandomVoiceJoin, originalHandler)
}

func processGuild(ctx context.Context, session *disruptor.Session, guild snowflake.ID) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if _, ok := session.Caches().Guild(guild); !ok {
		return nil // Skip if guild is not in cache
	}

	// Get available voice channels
	channels, err := getAvailableVoiceChannels(ctx, session, guild)
	if err != nil {
		return fmt.Errorf("failed to get channels for guild %s: %w", guild, err)
	}

	if len(channels) == 0 {
		return nil // No available channels, skip
	}

	// Select a random channel
	channelID := randomChannelID(channels)

	// Record voice connection attempt
	audioMetrics := metrics.NewAudioMetrics()

	if err := session.UpdateVoiceState(ctx, guild, &channelID); err != nil {
		audioMetrics.RecordVoiceConnectionAttempt(guild, false)
		return fmt.Errorf("failed to update voice state: %w", err)
	}

	audioMetrics.RecordVoiceConnectionAttempt(guild, true)
	return nil
}

func randomChannelID(channels []discord.GuildChannel) snowflake.ID {
	return channels[util.RandomInt(0, len(channels)-1)].ID()
}

func getAvailableVoiceChannels(ctx context.Context, session *disruptor.Session, guildID snowflake.ID) ([]discord.GuildChannel, error) {
	if session.Caches().GuildSoundboardSoundsLen(guildID) == 0 {
		return nil, fmt.Errorf("there are no soundboard sounds available")
	}

	// Record Discord API call metrics
	discordMetrics := metrics.NewDiscordAPIMetrics()
	timer := discordMetrics.NewDiscordAPITimer("/guilds/{guild.id}/channels", "GET")

	channels, err := session.Rest().GetGuildChannels(guildID, rest.WithCtx(ctx))

	statusCode := 200
	if err != nil {
		statusCode = 500 // Assume server error for failed requests
	}
	timer.Finish(statusCode)

	if err != nil {
		return nil, err
	}

	member, ok := session.Caches().Member(guildID, session.ID())
	if !ok {
		return nil, fmt.Errorf("bot is not a member of the guild %s", guildID)
	}

	filtered := make([]discord.GuildChannel, 0, len(channels))
	for _, channel := range channels {
		if channel.Type() != discord.ChannelTypeGuildVoice {
			continue
		}

		voiceChannel, ok := session.Caches().GuildVoiceChannel(channel.ID())
		if !ok {
			continue
		}

		permissions := session.Caches().MemberPermissionsInChannel(voiceChannel, member)
		if !util.HasVoicePermissions(permissions) {
			continue
		}

		if len(session.Caches().AudioChannelMembers(voiceChannel)) == 0 {
			continue
		}

		filtered = append(filtered, channel)
	}

	return filtered, nil
}

func getEligibleGuilds(ctx context.Context, db *bun.DB, interval time.Duration, chance float64) ([]models.Guild, error) {
	guilds := make([]models.Guild, 0)
	if err := db.NewSelect().Model(&guilds).Where("chance <= ? AND interval = ?", chance, interval).Scan(ctx, &guilds); err != nil {
		return nil, fmt.Errorf("failed to find eligible guilds: %w", err)
	}

	if len(guilds) == 0 {
		return nil, nil // No eligible guilds
	}

	return guilds, nil
}
