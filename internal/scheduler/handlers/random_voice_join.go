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
	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/internal/scheduler"
	"github.com/XanderD99/disruptor/internal/util"
)

const HandlerTypeRandomVoiceJoin = "random_voice_join"

func NewRandomVoiceJoinHandler(session *disruptor.Disruptor, db *bun.DB) scheduler.HandleFunc {
	registerHandlerSingleton(HandlerTypeRandomVoiceJoin, func() any {
		return newRandomVoiceJoinHandler(session, db)
	})

	cb, ok := getHandlerSingleton(HandlerTypeRandomVoiceJoin).(scheduler.HandleFunc)
	if !ok {
		panic(fmt.Sprintf("handler %s is not a scheduler.HandleFunc", HandlerTypeRandomVoiceJoin))
	}
	return cb
}

func newRandomVoiceJoinHandler(session *disruptor.Disruptor, db *bun.DB) scheduler.HandleFunc {
	return func(ctx context.Context) error {
		chance := util.RandomInt(0, 101) // Use float for better precision

		interval, ok := util.GetIntervalFromContext(ctx)
		if !ok {
			return fmt.Errorf("failed to get interval from context")
		}

		guilds, err := getEligibleGuilds(ctx, db, interval, chance)
		if err != nil {
			return fmt.Errorf("failed to find guilds: %w", err)
		}

		maxWorkers := int(math.Max(1, math.Sqrt(float64(len(guilds)))))

		return util.ProcessWithWorkerPool(ctx, guilds, maxWorkers, func(ctx context.Context, guild models.Guild) {
			if err := processGuild(ctx, session, guild); err != nil {
				session.Logger.ErrorContext(ctx, "Failed to process guild", slog.Any("guild.id", guild.ID), slog.Any("error", err))
			}
		})
	}
}

func processGuild(ctx context.Context, session *disruptor.Disruptor, guild models.Guild) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if _, ok := session.Caches.Guild(guild.ID); !ok {
		return nil // Skip if guild is not in cache
	}

	// Get available voice channels
	channelID, err := determineVoiceChannelID(ctx, session, guild)
	if err != nil {
		return fmt.Errorf("failed to get channels for guild %s: %w", guild.ID, err)
	}

	// Record voice connection attempt

	sound, err := util.GetRandomSound(session.Client, guild.ID)
	if err != nil {
		return fmt.Errorf("failed to get random sound: %w", err)
	}

	if err := util.PlaySound(ctx, session.Client, guild.ID, channelID, sound.URL()); err != nil {
		return fmt.Errorf("failed to play sound: %w", err)
	}

	return nil
}

func determineVoiceChannelID(ctx context.Context, session *disruptor.Disruptor, guild models.Guild) (snowflake.ID, error) {
	available, err := getAvailableVoiceChannels(ctx, session, guild)
	if err != nil {
		return 0, err
	}

	if len(available) == 0 {
		return 0, nil // No available channels
	}

	if len(available) == 1 {
		return available[0], nil // Only one available channel, return it
	}

	// Build weights map from guild.Channels, default to .5
	weights := make([]float64, len(available))
	for i, channelID := range available {
		weight := .5
		for _, ch := range guild.Channels {
			if ch.ID == channelID && ch.Weight > 0 {
				weight = ch.Weight
				break
			}
		}
		weights[i] = weight
	}

	// Weighted random selection
	total := 0.0
	for _, w := range weights {
		total += w
	}
	r := util.RandomFloat(0, total)
	for i, w := range weights {
		if r < w {
			return available[i], nil
		}
		r -= w
	}

	return available[0], nil // fallback
}

func getAvailableVoiceChannels(ctx context.Context, session *disruptor.Disruptor, guild models.Guild) ([]snowflake.ID, error) {
	if session.Caches.GuildSoundboardSoundsLen(guild.ID) == 0 {
		return nil, fmt.Errorf("there are no soundboard sounds available")
	}

	channels, err := session.Rest.GetGuildChannels(guild.ID, rest.WithCtx(ctx))
	if err != nil {
		return nil, err
	}

	member, ok := session.Caches.Member(guild.ID, session.ID())
	if !ok {
		return nil, fmt.Errorf("bot is not a member of the guild %s", guild.ID)
	}

	filtered := make([]snowflake.ID, 0)
	for _, channel := range channels {
		if channel.Type() != discord.ChannelTypeGuildVoice {
			continue
		}

		voiceChannel, ok := session.Caches.GuildVoiceChannel(channel.ID())
		if !ok {
			continue
		}

		permissions := session.Caches.MemberPermissionsInChannel(voiceChannel, member)
		if !util.HasVoicePermissions(permissions) {
			continue
		}

		if len(session.Caches.AudioChannelMembers(voiceChannel)) == 0 {
			continue
		}

		filtered = append(filtered, channel.ID())
	}

	return filtered, nil
}

func getEligibleGuilds(ctx context.Context, db *bun.DB, interval time.Duration, chance int) ([]models.Guild, error) {
	guilds := make([]models.Guild, 0)
	if err := db.NewSelect().Model(&guilds).Where("chance >= ? AND interval = ?", chance, interval).Relation("Channels").Scan(ctx); err != nil {
		return nil, fmt.Errorf("failed to find eligible guilds: %w", err)
	}

	if len(guilds) == 0 {
		return nil, nil // No eligible guilds
	}

	return guilds, nil
}
