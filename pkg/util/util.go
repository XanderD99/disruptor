package util

import (
	"context"
	"sync"

	"github.com/disgoorg/disgo/discord"
)

func HasVoicePermissions(permissions discord.Permissions) bool {
	return permissions.Has(discord.PermissionSpeak, discord.PermissionConnect, discord.PermissionViewChannel)
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
