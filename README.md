# Discord Disruptor 🎉

Welcome to **Discord Disruptor**, the bot that brings delightful chaos to your Discord server! This mischievous little bot randomly joins voice channels at unpredictable intervals to play surprise sound clips, keeping your community on their toes. Whether you're looking to add some spontaneous fun or just want to give your friends a good laugh (or jump scare), this bot has got you covered. Built with Go, it's fast, reliable, and ready to disrupt your voice channels in the most entertaining way possible! 🚀🔊

## What Does It Do? 🎭

**Discord Disruptor** is the ultimate chaos agent for your Discord server. Here's the fun:

- 🎲 **Random Voice Channel Invasions**: The bot sneaks into voice channels at completely random intervals
- 🔊 **Surprise Sound Drops**: Plays random sound clips to surprise (and possibly startle) your friends
- ⏰ **Unpredictable Timing**: You never know when it's coming next - that's the beauty of it!
- 🎵 **Sound Library**: Supports a collection of audio files for maximum variety and chaos
- 🎯 **Smart Channel Selection**: Intelligently picks active voice channels for maximum impact

## Features ✨

- **Tick-Tock Magic** ⏰: Advanced timer and ticker system with all the bells and whistles—intervals, jitter, grouping, and worker pools. Perfect for those unpredictable disruption schedules!
- **Audio Arsenal** 🎵: Supports various audio formats and maintains a library of sound clips for maximum chaos potential
- **Voice Channel Vigilance** 👁️: Monitors voice channels and picks the perfect moments to strike
- **Config Your Heart Out** 🛠️: Environment variable-based configuration so you can fine-tune your chaos levels
- **Dockerized Delight** 🐳: Optimized Dockerfile for a lean, mean, disruption machine
- **Metrics & Logs** 📊: Keep track of your successful disruptions with integrated metrics and structured logging
- **Modular Mayhem** 🧩: Add new sounds, commands, and disruption strategies with ease. The chaos never ends!

## Getting Started 🚀

### Prerequisites ✅

Before you unleash the chaos, make sure you have:

- [Go](https://golang.org/doc/install) (version 1.20 or later) 🐹
- [Docker](https://docs.docker.com/get-docker/) (because containers are cool) 🐳
- [Make](https://www.gnu.org/software/make/) (to make your life easier) 🛠️
- A Discord bot token (for the chaos to be official) 🤖
- Some audio files ready for disruption duty 🎵

### Installation 🛠️

1. Clone this beautiful mess of a repository:

   ```bash
   git clone https://github.com/your-repo/discord-disruptors.git
   cd discord-disruptors
   ```

2. Build the project (cue dramatic build music 🎵):

   ```bash
   make build
   ```

3. Set up your audio files in the designated sound directory 🎵

4. Configure your Discord bot token and let the chaos begin 🎉:

   ```bash
   ./output/bin/disruptor
   ```

### Docker Deployment 🐳

1. Build the Docker image (because containerized chaos is the best chaos):

   ```bash
   docker build -t discord-disruptor .
   ```

2. Run the container and watch the voice channel magic happen ✨:

   ```bash
   docker run -d --name discord-disruptor \
     -e CONFIG_TOKEN=your_discord_bot_token \
     -v /path/to/your/sounds:/app/sounds \
     discord-disruptor
   ```

## Configuration ⚙️

This bot is powered by environment variables that control the chaos levels. For more info have a look at [/internal/config](/internal/config/README.md)

## Development 🛠️

### Directory Structure 📂

Here's where the chaos is orchestrated:

- `cmd/`: Entry points for the disruption engine 🚪
- `internal/`: The brain of the operation (and the chaos algorithms) 🧠
- `sounds/`: Your audio arsenal for maximum disruption 🎵
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

Got a cool sound effect? Found a bug in the chaos algorithm? Want to make this bot even more delightfully disruptive? Contributions are welcome! Check out the [contribution guidelines](CONTRIBUTING.md).

Ideas for contributions:

- New sound effects and audio clips 🎵
- Better channel selection algorithms 🎯
- Additional disruption strategies 🎭
- Sound categorization and theming 🏷️

## License 📜

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgments 🙌

- [Discord API](https://discord.com/developers/docs/intro) for making voice channel chaos possible 💬
- [Go](https://golang.org/) for being awesome and fast 🐹
- [DCA](https://github.com/bwmarrin/dca) for Discord audio encoding magic 🎵
- Everyone who has been "disrupted" by this bot - you're the real heroes 🏆

And of course, you, for being here and ready to spread some harmless chaos. You're the real MVP! 🎉

---

**Warning**: Use responsibly! This bot is designed for fun and should be used in servers where everyone is okay with surprise audio interruptions. Always respect your community's preferences and server rules. Happy disrupting! 😄
