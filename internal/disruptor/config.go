package disruptor

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"
)

type Config struct {
	// 🔑 The bot token used to connect to Discord
	Token string `env:"TOKEN,required"`

	Sharding struct {
		// 🔢 Shard ID to use
		ShardIDs []int `env:"IDS" default:"0"`
		// 🔢 Total number of shards to use
		ShardCount int `env:"COUNT" default:"1"`
		// 🔢 Whether to enable autoscaling for shards
		Autoscaling bool `env:"AUTOSCALING" default:"false"`
	} `envPrefix:"SHARDING_"`
}

func (c Config) ToShardManagerOpts() []sharding.ConfigOpt {
	opts := []sharding.ConfigOpt{
		sharding.WithShardCount(c.Sharding.ShardCount),
		sharding.WithAutoScaling(c.Sharding.Autoscaling),
		sharding.WithGatewayConfigOpts(
			gateway.WithIntents(gateway.IntentGuilds, gateway.IntentGuildVoiceStates, gateway.IntentGuildExpressions, gateway.IntentGuildMembers),
			gateway.WithCompress(true),
			gateway.WithPresenceOpts(gateway.WithListeningActivity("to your soundboards")),
		),
	}

	if len(c.Sharding.ShardIDs) > 0 {
		opts = append(opts, sharding.WithShardIDs(c.Sharding.ShardIDs...))
	}

	return opts
}

func (c Config) ToSessionOpts(opts ...sharding.ConfigOpt) []bot.ConfigOpt {
	return []bot.ConfigOpt{
		bot.WithShardManagerConfigOpts(c.ToShardManagerOpts()...),
		bot.WithCacheConfigOpts(cache.WithCaches(cache.FlagsAll)),
	}
}
