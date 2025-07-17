# Disruptor 🎉

Welcome to **Disruptor**, the bot that brings delightful chaos to your Discord server! This mischievous little bot randomly joins voice channels at unpredictable intervals to play surprise sounds directly from your server's Discord soundboard, keeping your community on their toes. Whether you're looking to add some spontaneous fun or just want to give your friends a good laugh (or jump scare), this bot has got you covered. Built with Go, it's fast, reliable, and leverages Discord's native soundboard feature - no external sound storage required! 🚀🔊

## What Does It Do? 🎭

**Disruptor** is the ultimate chaos agent for your Discord server. Here's the fun:

- 🎲 **Random Voice Channel Invasions**: The bot sneaks into voice channels at completely random intervals
- 🔊 **Native Soundboard Integration**: Plays sounds directly from your server's Discord soundboard with no external storage
- ⏰ **Unpredictable Timing**: You never know when it's coming next - that's the beauty of it!
- 🎵 **Auto-Discovery**: Automatically uses all soundboard sounds available in your server
- 🎯 **Smart Channel Selection**: Intelligently picks active voice channels for maximum impact
- ⚙️ **Per-Guild Configuration**: Each server can customize disruption intervals and preferences

## Features ✨

- **Tick-Tock Magic** ⏰: Advanced timer and ticker system with all the bells and whistles—intervals, jitter, grouping, and worker pools. Perfect for those unpredictable disruption schedules!
- **Native Soundboard Support** 🎵: Leverages Discord's built-in soundboard feature - no external sound storage needed!
- **Simplified Architecture** 🔧: Sounds come directly from Discord, only guild settings are stored in the database
- **Voice Channel Vigilance** 👁️: Monitors voice channels and picks the perfect moments to strike
- **Config Your Heart Out** 🛠️: Environment variable-based configuration so you can fine-tune your chaos levels
- **Dockerized Delight** 🐳: Optimized Dockerfile for a lean, mean, disruption machine
- **Metrics & Logs** 📊: Keep track of your successful disruptions with integrated metrics and structured logging
- **Modular Mayhem** 🧩: Add new disruption strategies with ease. The chaos never ends!

## Getting Started 🚀

### Prerequisites ✅

Before you unleash the chaos, make sure you have:

- [Go](https://golang.org/doc/install) (version 1.20 or later) 🐹
- [Docker](https://docs.docker.com/get-docker/) (because containers are cool) 🐳
- [Make](https://www.gnu.org/software/make/) (to make your life easier) 🛠️
- A Discord bot token (for the chaos to be official) 🤖
- Discord soundboard sounds uploaded to your server 🎵
- Database for storing guild configuration (MongoDB recommended) 💾

### Installation 🛠️

1. Clone this beautiful mess of a repository:

   ```bash
   git clone https://github.com/XanderD99/disruptor.git
   cd disruptor
   ```

2. Build the project (cue dramatic build music 🎵):

   ```bash
   make build
   ```

3. Upload sounds to your Discord server's soundboard 🎵

4. Configure your Discord bot token and database connection 🎉:

   ```bash
   export CONFIG_TOKEN=your_discord_bot_token
   export CONFIG_DATABASE_URL=your_database_connection
   ./output/bin/disruptor
   ```

### Docker Deployment 🐳

1. Build the Docker image (because containerized chaos is the best chaos):

   ```bash
   docker build -t disruptor .
   ```

2. Run the container and watch the voice channel magic happen ✨:

   ```bash
   docker run -d --name disruptor \
     -e CONFIG_TOKEN=your_discord_bot_token \
     -e CONFIG_DATABASE_URL=your_database_connection \
     disruptor
   ```

## How It Works 🔧

The bot operates with a simplified architecture focused on Discord's native features:

1. **Soundboard Discovery**: Bot automatically detects all soundboard sounds available in your server
2. **Random Selection**: When it's time to disrupt, the bot randomly selects from available soundboard sounds
3. **Native Playback**: Sounds are played using Discord's native soundboard API for optimal quality
4. **Smart Timing**: Advanced algorithms determine the perfect moments to strike
5. **Guild Configuration**: Only server settings (intervals, preferences) are stored in the database

## Configuration ⚙️

This bot is powered by environment variables that control the chaos levels. Key configurations include:

- **Discord Bot Token**: Your bot's authentication token
- **Database Connection**: For storing guild configuration and preferences
- **Disruption Intervals**: How often the bot should strike
- **Channel Selection**: Which voice channels to target

For more info have a look at [/internal/config](/internal/config/README.md)

## Development 🛠️

### Directory Structure 📂

Here's where the chaos is orchestrated:

- `cmd/`: Entry points for the disruption engine 🚪
- `internal/`: The brain of the operation (and the chaos algorithms) 🧠
  - `commands/`: Slash commands for bot interaction 🎛️
  - `handlers/`: Event handlers for voice updates and soundboard integration 🎵
  - `models/`: Data models for guild configuration 📊
- `ci/`: Continuous integration and Docker wizardry 🧙‍♂️
- `output/`: Where the bot comes to life (compiled binaries and other goodies) 🎁

### Running Tests ✅

Want to make sure your chaos engine works perfectly? Run:

```bash
make test
```

### Linting ✨

Keep your chaos code clean and shiny:

```bash
make lint
```

## Contributing 🤝

Got ideas for better disruption strategies? Found a bug in the chaos algorithm? Want to make this bot even more delightfully disruptive? Contributions are welcome! Check out the [contribution guidelines](CONTRIBUTING.md).

Ideas for contributions:

- Better channel selection algorithms 🎯
- Additional disruption strategies 🎭
- Advanced guild configuration options ⚙️
- Performance optimizations 🚀
- Global soundboard support 🌍

## License 📜

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgments 🙌

- [Discord API](https://discord.com/developers/docs/intro) for making voice channel chaos and soundboard integration possible 💬
- [Disgo](https://github.com/disgoorg/disgo) for excellent Discord API bindings 🔗
- [Go](https://golang.org/) for being awesome and fast 🐹
- [Lavalink](https://github.com/freyacodes/Lavalink) for audio streaming capabilities 🎵
- Everyone who has been "disrupted" by this bot - you're the real heroes 🏆

And of course, you, for being here and ready to spread some harmless chaos. You're the real MVP! 🎉

---

**Warning**: Use responsibly! This bot is designed for fun and should be used in servers where everyone is okay with surprise audio interruptions. Always respect your community's preferences and server rules. Make sure to configure your soundboard sounds appropriately for your audience. Happy disrupting! 😄
