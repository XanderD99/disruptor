# Disruptor 🎉

Welcome to **Disruptor**, the bot that brings delightful chaos to your Discord server! This mischievous Go bot randomly joins voice channels at unpredictable intervals to play surprise sounds from your server's Discord soundboard. Whether you want spontaneous fun, a good laugh, or a jump scare, Disruptor delivers! Fast, reliable, and built on Discord's native soundboard API—no external sound storage required. 🚀🔊

---

## What Does It Do? 🎭

- 🎲 **Random Voice Channel Invasions**: Sneaks into voice channels at random intervals for maximum surprise.
- 🔊 **Native Soundboard Integration**: Plays sounds directly from your server's Discord soundboard.
- ⏰ **Unpredictable Timing**: You never know when it's coming next!
- 🎵 **Auto-Discovery**: Automatically uses all soundboard sounds available in your server.
- 🎯 **Smart Channel Selection**: Picks active voice channels using weighted random selection for maximum impact.
- ⚙️ **Per-Guild Configuration**: Each server can customize disruption intervals and preferences.
- 🏃‍♂️ **Bun ORM & SQLite**: Fast, modern database operations with automatic schema management.
- 🛠️ **Environment Variable Config**: Fine-tune chaos levels with simple env vars.
- 🐳 **Docker-Ready**: Optimized Dockerfile for easy deployment.
- 📊 **Metrics & Logs**: Track disruptions with integrated metrics and structured logging.
- 🧩 **Modular Mayhem**: Easily add new disruption strategies.
- 🕵️ **Voice Channel Vigilance**: Monitors voice channels and picks the perfect moments to strike.
- ⚖️ **Weighted Channel Selection**: Set custom weights for voice channels to control disruption probability.
- 🧑‍💻 **Slash Commands**: Control the bot with Discord slash commands (`/play`, `/interval`, `/chance`, `/disconnect`, `/next`, `/weight`).
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

### 🏃‍♂️ Quick Start (10 minutes)

Want to get up and running immediately? Follow the **[Quick Start Guide](docs/QUICKSTART.md)** for the fastest setup path.

### 📚 Complete Setup Guides

Choose the guide that matches your experience level: 🎯

| Guide | Audience | Time | Description |
|-------|----------|------|-------------|
| **[Quick Start](docs/QUICKSTART.md)** | Everyone 👥 | 10 min ⚡ | Fastest path to a working bot 🚀 |
| **[Discord Setup](docs/DISCORD_SETUP.md)** | New to Discord bots 🆕 | 15 min 📝 | Step-by-step Discord bot creation 🤖 |
| **[Installation](docs/INSTALLATION.md)** | All platforms 🌐 | 20 min 🔧 | Detailed platform-specific installation 💻 |
| **[Configuration](docs/CONFIGURATION.md)** | Advanced users 🤓 | 30 min ⚙️ | Complete configuration reference 📖 |
| **[Deployment](docs/DEPLOYMENT.md)** | Production 🏭 | 45 min 🚀 | Production deployment strategies 🐳 |

### ⚡ Minimal Setup

If you just want to test locally (the lazy way 😎):

```bash
# 1. Create Discord bot (get token from Discord Developer Portal)
# 2. Clone and build
git clone https://github.com/XanderD99/disruptor.git
cd disruptor && go mod download && make build

# 3. Run with your token
export CONFIG_TOKEN="your_discord_bot_token"
./output/bin/disruptor
```

**Need help? 🤔** Start with the [Discord Setup Guide](docs/DISCORD_SETUP.md) if you're new to Discord bots.

📚 **[Complete Documentation Index](docs/README.md)** - Find the right guide for your needs! 🗺️

---

## Docker Deployment 🐳

### Quick Docker Setup 🐋

```bash
# Build image 🏗️
make docker-build

# Run with environment variables 🌍
docker run -d --name disruptor \
  -e CONFIG_TOKEN="your_discord_bot_token" \
  -e CONFIG_DATABASE_DSN="file:/data/disruptor.db?cache=shared" \
  -v $(pwd)/data:/data \
  disruptor:latest
```

### Docker Compose 🎼

```yaml
# docker-compose.yml 📝
version: '3.8'
services:
  disruptor:
    build: .
    environment:
      - CONFIG_TOKEN=${DISCORD_TOKEN}  # Set in .env file
      - CONFIG_DATABASE_DSN=file:/data/disruptor.db?cache=shared
    volumes:
      - ./data:/data
    restart: unless-stopped
```

```bash
echo "DISCORD_TOKEN=your_bot_token" > .env  # 🔑 Secret sauce!
docker-compose up -d  # 🚀 Blast off!
```

📖 **Production Deployment**: See [Deployment Guide](docs/DEPLOYMENT.md) for advanced deployment scenarios! 🏭✨

---

## How It Works 🔧

- **Soundboard Discovery** 🔍: Finds all soundboard sounds in your server.
- **Random Selection** 🎲: Picks a sound and a channel at random.
- **Native Playback** 🔊: Uses Discord's soundboard API for high-quality audio.
- **Smart Timing** 🧠: Advanced scheduler determines when to disrupt.
- **Guild Configuration** ⚙️: Stores only server settings (intervals, preferences) in the database.
- **Slash Commands** ⚡: Control disruption, intervals, chance, and more directly from Discord.

---

## Configuration ⚙️

### Essential Configuration 🔧

**Required (or it won't work! 😱):**

- `CONFIG_TOKEN`: Your Discord bot token (get from [Discord Developer Portal](https://discord.com/developers/applications)) 🎟️

**Recommended (for a better experience 💯):**

- `CONFIG_DATABASE_DSN`: Database location (`file:./disruptor.db?cache=shared` for persistent storage) 🗄️
- `CONFIG_LOGGING_LEVEL`: Log verbosity (`debug`, `info`, `warn`, `error`) 📝

### Configuration Examples 💡

```bash
# Development (for testing & breaking things 🧪)
export CONFIG_TOKEN="your_bot_token"
export CONFIG_DATABASE_DSN="file::memory:?cache=shared"  # In-memory (poof! gone when you restart 💨)
export CONFIG_LOGGING_LEVEL="debug"

# Production (for the real deal 🎯)
export CONFIG_TOKEN="your_bot_token"
export CONFIG_DATABASE_DSN="file:./disruptor.db?cache=shared"  # Persistent (keeps your data safe 🛡️)
export CONFIG_LOGGING_LEVEL="info"
```

📖 **Complete Reference**: See [Configuration Guide](docs/CONFIGURATION.md) for all options and examples! 📚✨

---

## Database 🗃️

- **SQLite + Bun ORM** 💪: Zero config, fast, reliable, portable.
- **Options** (pick your flavor! 🍦):
  - In-memory: `CONFIG_DATABASE_DSN=file::memory:?cache=shared` ⚡ (fast but forgetful!)
  - File-based: `CONFIG_DATABASE_DSN=file:./disruptor.db?cache=shared` 💾 (remembers everything!)
  - Custom path: `CONFIG_DATABASE_DSN=file:/path/to/your/database.db?cache=shared` 🗂️ (your way!)
- **Automatic Schema Management** 🤖: Tables and migrations handled on startup.

---

## Development 🛠️

### Directory Structure 📂

- `cmd/` 🚪: Application entrypoints
- `internal/` 🔒: Core application logic
  - `commands/` ⚡: Discord slash commands (`/play`, `/interval`, `/chance`, `/disconnect`, `/next`)
  - `handlers/` 🎯: Discord event handlers
  - `models/` 🗄️: Database models (Bun ORM)
  - `scheduler/` ⏰: Audio scheduling and timing logic
  - `metrics/` 📊: Prometheus metrics and monitoring
  - `util/` 🛠️: Shared utilities
- `pkg/` 📦: Reusable packages (logging, processes)
- `ci/` 🏭: CI/CD configuration and Docker files
- `docs/` 📚: Documentation
- `output/` 🎯: Compiled binaries

### Build and Test 🧪

```bash
# Download dependencies 📦
go mod download

# Build application 🔨
make build

# Run tests (currently no tests implemented 😅)
make test

# Static analysis and formatting 🧹
go vet ./...
go fmt ./...
```

**Note** ⚠️: Do not use `make lint` - there are known configuration issues with golangci-lint. Use `go vet ./...` instead.

### Development Setup 💻

```bash
# Clone repository 📥
git clone https://github.com/XanderD99/disruptor.git
cd disruptor

# Install system dependencies (Ubuntu/Debian) 🔧
sudo apt install -y libopus-dev pkg-config

# Build and run in development mode 🚀
go mod download
make build
export CONFIG_TOKEN="your_dev_bot_token"
export CONFIG_LOGGING_LEVEL="debug"
./output/bin/disruptor
```

📖 **Contributing**: See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines! 🤝✨

---

## Slash Commands 🎛️

- `/play` 🎵 — Play a soundboard sound immediately
- `/interval` ⏱️ — Set disruption interval per guild
- `/chance` 🎲 — Set disruption chance per guild
- `/weight` ⚖️ — Set channel selection weight (0-100, higher = more likely to be chosen)
- `/disconnect` 🛑 — Instantly stop disruptions
- `/next` 🔮 — Preview next scheduled disruption

---

## Weighted Channel Selection ⚖️

Disruptor now supports **weighted channel selection**, allowing you to control which voice channels are more likely to be chosen for disruptions! 🎯✨

### How It Works 🔧

- Each voice channel has a **weight** that determines its selection probability 📊
- **Default weight**: 50 (balanced selection) ⚖️
- **Range**: 0-100 (0 = never selected 🚫, 100 = maximum priority 🌟)
- **Smart selection**: Channels with higher weights are proportionally more likely to be chosen 🧠

### Managing Weights 🎚️

```bash
# Set a channel's weight 🎯
/weight channel:#general-voice weight:75

# Check current weight (omit weight parameter) 👀
/weight channel:#general-voice

# High priority channel (disrupted more often) 🔥
/weight channel:#chaos-room weight:90

# Low priority channel (disrupted less often) 🤫
/weight channel:#quiet-study weight:10

# Exclude channel from disruptions 🚫
/weight channel:#meeting-room weight:0
```

### Examples 💡

If you have 3 channels with these weights:

- 🎵 **Music Room**: Weight 80 (party time! 🎉)
- 🗣️ **General Chat**: Weight 50 (default balance ⚖️)
- 📚 **Study Hall**: Weight 20 (keep it quiet 🤫)

**Music Room** will be disrupted ~53% of the time, **General Chat** ~33%, and **Study Hall** ~13%. Math! 🧮✨

### Use Cases 🎪

- **Priority channels** 🌟: Boost weight for channels where disruptions are most welcome
- **Protected spaces** 🛡️: Lower weight for serious discussion channels
- **Event management** 🎊: Temporarily adjust weights for special events
- **Community preferences** 🗳️: Let server members vote on channel disruption preferences

---

## Troubleshooting 🔧

### Common Issues 🤕

| Issue | Quick Fix |
|-------|-----------|
| "CONFIG_TOKEN is not set" 😱 | `export CONFIG_TOKEN="your_bot_token"` 🔑 |
| Bot won't connect 😵 | Verify token in Discord Developer Portal 🔍 |
| No slash commands 😭 | Wait 1 hour for sync, or re-invite bot ⏰ |
| Can't join voice 🚫 | Check bot has Connect + Speak permissions 🎤 |
| No sounds playing 🔇 | Upload sounds to server soundboard 🎵 |

### Getting Help 🆘

- 🚀 **Quick fixes**: [Troubleshooting Guide](docs/TROUBLESHOOTING.md) (coming soon! 🔜)
- 📖 **Setup help**: [Setup Guide](docs/SETUP.md) (your new best friend 🤗)
- ⚙️ **Config issues**: [Configuration Guide](docs/CONFIGURATION.md) (all the knobs and dials 🎛️)
- 🐛 **Bug reports**: [GitHub Issues](https://github.com/XanderD99/disruptor/issues) (oops, we broke something! 😅)
- 💬 **Questions**: [GitHub Discussions](https://github.com/XanderD99/disruptor/discussions) (let's chat! ☕)

---

## Contributing 🤝

We welcome contributions! Here are some areas where help is appreciated: 🙋‍♀️🙋‍♂️

- 🎯 **Smart channel selection algorithms** - Better logic for choosing voice channels (make it smarter! 🧠)
- 🎭 **New disruption strategies** - Creative ways to surprise users (chaos innovation! 💡)
- ⚙️ **Advanced configuration options** - More customization features (all the knobs! 🎛️)
- 🚀 **Performance optimizations** - Memory and CPU improvements (faster chaos! ⚡)
- 🌍 **Global soundboard support** - Cross-server sound sharing (world domination! 🌎)
- 📝 **Documentation improvements** - Better guides and examples (make it clearer! 📖)
- 🧪 **Test coverage** - Unit and integration tests (break it properly! 🔨)

📖 **Read more**: [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines! 🤝✨

---

## License 📜

MIT License. See [LICENSE](LICENSE).

---

## Acknowledgments 🙌

- [Discord API](https://discord.com/developers/docs/intro) 🤖 (the magic behind the curtain)
- [Disgo](https://github.com/disgoorg/disgo) ⚡ (our Discord superhero library)
- [Bun](https://bun.uptrace.dev/) 🗄️ (database wizardry)
- [Go](https://golang.org/) 🐹 (gopher power!)
- [SQLite](https://www.sqlite.org/) 📱 (tiny but mighty database)
- Everyone who's been "disrupted" by this bot—you're the real MVP! 🏆💖

---

**Warning** ⚠️: Use responsibly! This bot is for fun and should only be used in servers where surprise audio is welcome. Always respect your community's preferences and server rules. Happy disrupting! 🎉😈
