package util

import (
	"context"
	"time"
)

type intervalKey struct{}

func AddIntervalToContext(ctx context.Context, interval time.Duration) context.Context {
	return context.WithValue(ctx, intervalKey{}, interval)
}

func GetIntervalFromContext(ctx context.Context) (time.Duration, bool) {
	interval, ok := ctx.Value(intervalKey{}).(time.Duration)
	return interval, ok
}
