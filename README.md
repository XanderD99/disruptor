# Disruptor 🎉

Welcome to **Disruptor**, the bot that brings delightful chaos to your Discord server! This mischievous Go bot randomly joins voice channels at unpredictable intervals to play surprise sounds from your server's Discord soundboard. Whether you want spontaneous fun, a good laugh, or a jump scare, Disruptor delivers! Fast, reliable, and built on Discord's native soundboard API—no external sound storage required. 🚀🔊

---

## What Does It Do? 🎭

- 🎲 **Random Voice Channel Invasions**: Sneaks into voice channels at random intervals for maximum surprise.
- 🔊 **Native Soundboard Integration**: Plays sounds directly from your server's Discord soundboard.
- ⏰ **Unpredictable Timing**: You never know when it's coming next!
- 🎵 **Auto-Discovery**: Automatically uses all soundboard sounds available in your server.
- 🎯 **Smart Channel Selection**: Picks active voice channels for maximum impact.
- ⚙️ **Per-Guild Configuration**: Each server can customize disruption intervals and preferences.
- 🏃‍♂️ **Bun ORM & SQLite**: Fast, modern database operations with automatic schema management.
- 🛠️ **Environment Variable Config**: Fine-tune chaos levels with simple env vars.
- 🐳 **Docker-Ready**: Optimized Dockerfile for easy deployment.
- 📊 **Metrics & Logs**: Track disruptions with integrated metrics and structured logging.
- 🧩 **Modular Mayhem**: Easily add new disruption strategies.
- 🕵️ **Voice Channel Vigilance**: Monitors voice channels and picks the perfect moments to strike.
- 🧑‍💻 **Slash Commands**: Control the bot with Discord slash commands (`/play`, `/interval`, `/chance`, `/disconnect`, `/next`).
- 🎚️ **Interval & Chance Control**: Adjust how often and how likely disruptions are per guild.
- 🛑 **Manual Disconnect**: Instantly stop disruptions with a command.
- 🔄 **Next Disruption Preview**: See when the next chaos event is scheduled.
- 🧠 **Advanced Scheduler**: Worker pools, jitter, grouping, and more for unpredictable disruption schedules.
- 🏷️ **Structured Logging**: Debug and monitor with clear, contextual logs.
- 🏗️ **Automatic Schema Migration**: Database tables and schema managed on startup.
- 🧪 **Test Infrastructure**: Ready for tests (no unit tests yet, but infra is set up).
- 🛡️ **Safe Defaults**: In-memory DB for dev, file-based for production.

---

## Getting Started 🚀

### Prerequisites ✅

- [Go](https://golang.org/doc/install) **1.24+** 🐹
- [Docker](https://docs.docker.com/get-docker/) 🐳
- [Make](https://www.gnu.org/software/make/) 🛠️
- Discord bot token 🤖
- Discord soundboard sounds uploaded 🎵
- SQLite database (auto-created) 💾

### Installation 🛠️

1. Clone the repo:

   ```bash
   git clone https://github.com/XanderD99/disruptor.git
   cd disruptor
   ```

2. Download dependencies:

   ```bash
   go mod download
   ```

3. Build the project:

   ```bash
   make build
   ```

4. Configure your Discord bot token and (optionally) your SQLite database path:

   ```bash
   export CONFIG_TOKEN=your_discord_bot_token
   export CONFIG_DATABASE_DSN=file:./disruptor.db?cache=shared
   ./output/bin/disruptor
   ```

---

## Docker Deployment 🐳

1. Build the Docker image:

   ```bash
   make docker-build
   # OR
   docker build --target final -t disruptor:latest -f ./ci/Dockerfile .
   ```

2. Run the container:

   ```bash
   docker run -d --name disruptor \
     -e CONFIG_TOKEN=your_discord_bot_token \
     -e CONFIG_DATABASE_DSN=file:./disruptor.db?cache=shared \
     -v /host/path/to/data:/app/data \
     disruptor
   ```

   > **Note**: Defaults to in-memory SQLite. For persistence, mount a volume and use a file-based DSN.

---

## How It Works 🔧

- **Soundboard Discovery**: Finds all soundboard sounds in your server.
- **Random Selection**: Picks a sound and a channel at random.
- **Native Playback**: Uses Discord's soundboard API for high-quality audio.
- **Smart Timing**: Advanced scheduler determines when to disrupt.
- **Guild Configuration**: Stores only server settings (intervals, preferences) in the database.
- **Slash Commands**: Control disruption, intervals, chance, and more directly from Discord.

---

## Configuration ⚙️

All settings via environment variables:

- `CONFIG_TOKEN` (required): Discord bot token
- `CONFIG_DATABASE_DSN`: SQLite DSN (`file:./disruptor.db?cache=shared` for file-based, `file::memory:?cache=shared` for in-memory)
- `CONFIG_LOGGING_LEVEL`: Log verbosity (`debug`, `info`, `warn`, `error`)
- See `configs/.env.example` for full list

---

## Database 🗃️

- **SQLite + Bun ORM**: Zero config, fast, reliable, portable.
- **Options**:
  - In-memory: `CONFIG_DATABASE_DSN=file::memory:?cache=shared`
  - File-based: `CONFIG_DATABASE_DSN=file:./disruptor.db?cache=shared`
  - Custom path: `CONFIG_DATABASE_DSN=file:/path/to/your/database.db?cache=shared`
- **Automatic Schema Management**: Tables and migrations handled on startup.

---

## Development 🛠️

### Directory Structure 📂

- `cmd/`: Entrypoints
- `internal/`: Core logic
  - `commands/`: Slash commands (`play`, `interval`, `chance`, `disconnect`, `next`)
  - `handlers/`: Discord event handlers
  - `models/`: Database models
  - `scheduler/`: Audio scheduling logic
  - `metrics/`: Metrics and monitoring
  - `util/`: Utilities
- `ci/`: CI/CD and Docker
- `output/`: Compiled binaries

### Build, Test, and Lint 🧪

- **Build**: `make build`
- **Test**: `make test` (no unit tests yet)
- **Static Analysis**: `go vet ./...`
- **Format**: `go fmt ./...`
- **Lint**: *Do not use `make lint`—see instructions above*

---

## Slash Commands 🎛️

- `/play` — Play a soundboard sound immediately
- `/interval` — Set disruption interval per guild
- `/chance` — Set disruption chance per guild
- `/disconnect` — Instantly stop disruptions
- `/next` — Preview next scheduled disruption

---

## Contributing 🤝

Ideas for contributions:

- Smarter channel selection algorithms 🎯
- New disruption strategies 🎭
- Advanced guild config options ⚙️
- Performance optimizations 🚀
- Global soundboard support 🌍

See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License 📜

MIT License. See [LICENSE](LICENSE).

---

## Acknowledgments 🙌

- [Discord API](https://discord.com/developers/docs/intro)
- [Disgo](https://github.com/disgoorg/disgo)
- [Bun](https://bun.uptrace.dev/)
- [Go](https://golang.org/)
- [SQLite](https://www.sqlite.org/)
- Everyone who's been "disrupted" by this bot—you're the real MVP! 🏆

---

**Warning**: Use responsibly! This bot is for fun and should only be used in servers where surprise audio is welcome. Always respect your community's preferences and server rules. Happy disrupting!
