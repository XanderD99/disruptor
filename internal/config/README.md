# Environment Variables

## Config

 - `CONFIG_TOKEN` (**required**) - 🔑 The bot token used to connect to Discord
 - `CONFIG_SHARD_ID` (default: `0`) - 🔢 Shard ID to use, 0 for automatic assignment
 - `CONFIG_SHARD_COUNT` (default: `1`) - 🔢 Total number of shards to use, 0 for automatic calculation
 - `CONFIG_LOGGING_LEVEL` (default: `debug`) - 📜 Log level for the bot (e.g., debug, info, warn, error)
 - `CONFIG_LOGGING_PRETTY` (default: `true`) - ✨ Enable pretty-printed logs for human readability
 - `CONFIG_LOGGING_COLORS` (default: `true`) - 🌈 Add colors to logs for better visibility
 - `CONFIG_LOGGING_SOURCE` (default: `false`) - 🗂️ Include short file paths in log messages for debugging
 - `CONFIG_METRICS_SHUTDOWN_DURATION` (default: `15s`) - ⏳ How long to wait before shutting down the metrics server
 - `CONFIG_METRICS_PORT` (default: `9090`) - 📊 Port where the metrics server will be available
 - `CONFIG_API_BASE_URL` (default: `http://localhost:1337/api`) - 🌐 The base URL for the Strapi API
 - `CONFIG_API_AUTH_SCHEME` (default: `Bearer`) - 🔐 Authentication scheme for the Strapi
 - `CONFIG_API_AUTH_TOKEN` - 🔑 Authentication token for accessing the Strapi API
 - `CONFIG_API_DEBUG` (default: `false`) - 🛠️ Enable debug logging for Strapi API requests
 - `CONFIG_API_RETRY_COUNT` (default: `3`) - 🔁 Number of retries for failed requests
 - `CONFIG_API_RETRY_WAIT_TIME` (default: `1s`) - ⏳ Time to wait between retries
 - `CONFIG_API_RETRY_MAX_WAIT_TIME` (default: `5s`) - ⏳ Maximum time to wait for retries
 - `CONFIG_LAVALINK_NODENAME` (default: `disruptor`) - 🏷️ Name of the Lavalink node (must be unique)
 - `CONFIG_LAVALINK_NODEADDRESS` (default: `localhost:2333`) - 🌐 Lavalink server address (e.g., localhost:2333)
 - `CONFIG_LAVALINK_NODEPASSWORD` - 🔑 Lavalink server password
 - `CONFIG_LAVALINK_NODESECURE` (default: `false`) - 🔒 Use secure connection (wss)
