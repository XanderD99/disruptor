# Environment Variables

## Config

 - `CONFIG_TOKEN` (**required**) - 🔑 The bot token used to connect to Discord
 - `CONFIG_SHARDING_IDS` (comma-separated, default: `0`) - 🔢 Shard ID to use
 - `CONFIG_SHARDING_COUNT` (default: `1`) - 🔢 Total number of shards to use
 - `CONFIG_SHARDING_AUTOSCALING` (default: `false`) - 🔢 Whether to enable autoscaling for shards
 - `CONFIG_LOGGING_LEVEL` (default: `debug`) - 📜 Log level for the bot (e.g., debug, info, warn, error)
 - `CONFIG_LOGGING_PRETTY` (default: `true`) - ✨ Enable pretty-printed logs for human readability
 - `CONFIG_LOGGING_COLORS` (default: `true`) - 🌈 Add colors to logs for better visibility
 - `CONFIG_LOGGING_SOURCE` (default: `false`) - 🗂️ Include short file paths in log messages for debugging
 - `CONFIG_LOGGING_DISCORD_WEBHOOK` - 📡 Discord webhook URL for sending log messages
 - `CONFIG_LOGGING_DISCORD_MIN_LEVEL` (default: `4`) - 📉 Minimum log level for Discord messages, defaults to warn level
 - `CONFIG_LOGGING_DISCORD_SYNC` (default: `false`) - 📦 Whether to wait for Discord messages to be sent before continuing
 - `CONFIG_METRICS_SHUTDOWN_DURATION` (default: `15s`) - ⏳ How long to wait before shutting down the metrics server
 - `CONFIG_METRICS_PORT` (default: `9090`) - 📊 Port where the metrics server will be available
 - `CONFIG_LAVALINK_NODENAME` (default: `disruptor`) - 🏷️ Name of the Lavalink node (must be unique)
 - `CONFIG_LAVALINK_NODEADDRESS` (default: `localhost:2333`) - 🌐 Lavalink server address (e.g., localhost:2333)
 - `CONFIG_LAVALINK_NODEPASSWORD` - 🔑 Lavalink server password
 - `CONFIG_LAVALINK_NODESECURE` (default: `false`) - 🔒 Use secure connection (wss)
 - `CONFIG_DATABASE_TYPE` (default: `sqlite`) - 🔗 Database type to use
 - `CONFIG_DATABASE_DSN` (default: `file::memory:?cache=shared`) - 🔗 Database connection string
