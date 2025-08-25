package util

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"

	"github.com/XanderD99/disruptor/internal/ffmpeg"
)

func HasVoicePermissions(permissions discord.Permissions) bool {
	return permissions.Has(discord.PermissionSpeak, discord.PermissionConnect, discord.PermissionViewChannel)
}

func GetRandomSound(client bot.Client, guildID snowflake.ID) (discord.SoundboardSound, error) {
	sounds := make([]discord.SoundboardSound, 0)
	client.Caches().GuildSoundboardSoundsForEach(guildID, func(soundboardSound discord.SoundboardSound) {
		sounds = append(sounds, soundboardSound)
	})

	if len(sounds) == 0 {
		return discord.SoundboardSound{}, fmt.Errorf("no soundboard sounds available")
	}

	index := RandomInt(0, len(sounds)-1)
	return sounds[index], nil
}

func PlaySound(ctx context.Context, client bot.Client, guildID, channelID snowflake.ID, url string) error {
	conn := client.VoiceManager().CreateConn(guildID)

	if err := conn.Open(ctx, channelID, false, true); err != nil {
		return fmt.Errorf("error connecting to voice channel: %w", err)
	}

	cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer conn.Close(cleanupCtx)

	rs, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error opening sound URL: %w", err)
	}
	defer rs.Body.Close()

	// Stream through ffmpeg to get Opus frames
	opusProvider, err := ffmpeg.New(ctx, rs.Body)
	if err != nil {
		return fmt.Errorf("error creating opus provider: %w", err)
	}
	defer opusProvider.Close()

	conn.SetOpusFrameProvider(opusProvider)

	if err := opusProvider.Wait(); err != nil {
		return fmt.Errorf("error waiting for opus provider: %w", err)
	}

	return nil
}

func ProcessWithWorkerPool[T any](
	ctx context.Context,
	items []T,
	maxWorkers int,
	process func(context.Context, T),
) error {
	var wg sync.WaitGroup
	workers := make(chan struct{}, maxWorkers)

	for _, item := range items {
		wg.Add(1)
		select {
		case workers <- struct{}{}:
			go func(it T) {
				defer func() {
					<-workers
					wg.Done()
				}()
				process(ctx, it)
			}(item)
		case <-ctx.Done():
			wg.Done()
			return ctx.Err()
		}
	}

	wg.Wait()
	return nil
}
