# Quick Start Guide 🚀

Get Disruptor up and running in under 10 minutes! This guide covers the fastest path from zero to a working Discord bot.

## Prerequisites ✅

Before starting, ensure you have:
- [Go 1.24+](https://golang.org/dl/) installed
- [Git](https://git-scm.com/) installed
- A Discord account

## Step 1: Create Discord Bot (3 minutes) 🤖

### 1.1 Create Discord Application
1. Go to https://discord.com/developers/applications
2. Click "New Application"
3. Name it "My Disruptor Bot"
4. Click "Create"

### 1.2 Create Bot User
1. Go to "Bot" section in left sidebar
2. Click "Add Bot"
3. Copy the bot token (keep it safe!)

### 1.3 Set Bot Permissions
1. Go to "OAuth2" → "URL Generator"
2. Select scopes: `bot` and `applications.commands`
3. Select permissions:
   - ✅ View Channels
   - ✅ Connect
   - ✅ Speak
   - ✅ Use Voice Activity
   - ✅ Use Slash Commands
4. Copy the generated URL

### 1.4 Invite Bot to Server
1. Open the URL from step 1.3
2. Select your server
3. Click "Authorize"

## Step 2: Install Disruptor (2 minutes) 💻

### Option A: Quick Install (Linux/macOS)
```bash
# Install dependencies (Ubuntu/Debian)
sudo apt update && sudo apt install -y git make pkg-config libopus-dev golang-go

# Clone and build
git clone https://github.com/XanderD99/disruptor.git
cd disruptor
go mod download
make build
```

### Option B: Docker (Any OS)
```bash
git clone https://github.com/XanderD99/disruptor.git
cd disruptor
make docker-build
```

## Step 3: Add Soundboard Sounds (1 minute) 🎵

1. Go to your Discord server
2. Open Server Settings → Soundboard
3. Upload at least one sound file (.mp3 or .ogg)
4. Name your sound appropriately

## Step 4: Configure and Run (2 minutes) ⚙️

### Option A: Direct Run
```bash
# Set your bot token (replace with your actual token)
export CONFIG_TOKEN="your_discord_bot_token_here"

# Run the bot
./output/bin/disruptor
```

### Option B: Docker Run
```bash
docker run -d --name disruptor \
  -e CONFIG_TOKEN="your_discord_bot_token_here" \
  disruptor:latest
```

## Step 5: Test the Bot (1 minute) 🧪

### 5.1 Verify Bot is Online
- Check your Discord server - the bot should show as online
- Look for the green dot next to the bot's name

### 5.2 Test Slash Commands
In your Discord server, type `/` and you should see Disruptor commands:
- `/play` - Play a soundboard sound immediately  
- `/interval` - Set disruption interval
- `/chance` - Set disruption chance
- `/disconnect` - Stop disruptions
- `/next` - See next scheduled disruption

### 5.3 Join Voice Channel and Test
1. Join a voice channel in your server
2. Type `/play` and select a sound
3. The bot should join your channel and play the sound

## Success! 🎉

Your Disruptor bot is now running and ready to cause delightful chaos!

---

## What's Next?

### Customize Your Bot
- **Adjust intervals**: Use `/interval 30` to set 30-minute disruption intervals
- **Set chances**: Use `/chance 0.8` to set 80% chance of disruption
- **Add more sounds**: Upload more soundboard sounds for variety

### Production Setup
For running the bot 24/7, see these guides:
- [Deployment Guide](DEPLOYMENT.md) - Production deployment options
- [Configuration Guide](CONFIGURATION.md) - Advanced configuration
- [Installation Guide](INSTALLATION.md) - Platform-specific setup

### Advanced Features
- **Database persistence**: Set `CONFIG_DATABASE_DSN=file:./disruptor.db?cache=shared`
- **Logging**: Configure logging levels and Discord webhook notifications
- **Metrics**: Enable Prometheus metrics for monitoring

---

## Common Issues & Quick Fixes

### Bot Won't Connect
```bash
# Check token is set correctly
echo $CONFIG_TOKEN

# Verify token format (should start with letters/numbers)
# If using .env file:
source .env
./output/bin/disruptor
```

### No Slash Commands
- Wait up to 1 hour for commands to sync globally
- Try kicking and re-inviting the bot
- Ensure `applications.commands` scope was granted

### Bot Can't Join Voice
- Check bot has "Connect" and "Speak" permissions
- Verify voice channel isn't full
- Ensure bot has access to the voice channel

### No Sounds Playing
- Upload at least one sound to server soundboard
- Check sounds are under 5.2MB
- Verify bot has soundboard permissions

---

## One-Command Setup Scripts

### Linux/macOS Setup Script
```bash
#!/bin/bash
# Save as setup.sh and run: chmod +x setup.sh && ./setup.sh

echo "🚀 Disruptor Quick Setup"

# Install dependencies
if command -v apt >/dev/null; then
    sudo apt update && sudo apt install -y git make pkg-config libopus-dev golang-go
elif command -v brew >/dev/null; then
    brew install go git opus pkg-config
fi

# Clone and build
git clone https://github.com/XanderD99/disruptor.git
cd disruptor
go mod download
make build

echo "✅ Setup complete!"
echo "Next steps:"
echo "1. Set your bot token: export CONFIG_TOKEN='your_token'"
echo "2. Run the bot: ./output/bin/disruptor"
```

### Docker Compose Setup
```yaml
# Save as docker-compose.yml
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
# Create .env file
echo "DISCORD_TOKEN=your_discord_bot_token_here" > .env

# Run
docker-compose up -d
```

---

## Need Help?

### Quick Links
- 🔧 [Installation Issues](INSTALLATION.md#troubleshooting-installation)
- ⚙️ [Configuration Problems](CONFIGURATION.md#configuration-validation)  
- 🤖 [Discord Setup Help](DISCORD_SETUP.md#common-issues)
- 🐛 [Troubleshooting Guide](TROUBLESHOOTING.md)

### Support Channels
- 📖 **Documentation**: Check `/docs/` folder
- 🐛 **Bug Reports**: Open GitHub issue
- 💬 **Questions**: Use GitHub Discussions
- 🚨 **Urgent**: Check troubleshooting guide first

---

## Configuration Summary

### Minimal Configuration
```bash
export CONFIG_TOKEN="your_discord_bot_token"
./output/bin/disruptor
```

### Recommended Configuration
```bash
export CONFIG_TOKEN="your_discord_bot_token"
export CONFIG_DATABASE_DSN="file:./disruptor.db?cache=shared"
export CONFIG_LOGGING_LEVEL="info"
./output/bin/disruptor
```

### Production Configuration
```bash
export CONFIG_TOKEN="your_discord_bot_token"
export CONFIG_DATABASE_DSN="file:/opt/disruptor/data/bot.db?cache=shared"
export CONFIG_LOGGING_LEVEL="info"
export CONFIG_LOGGING_PRETTY="false"
export CONFIG_METRICS_PORT="9090"
./output/bin/disruptor
```

---

**Happy disrupting!** 🎭 Your bot is ready to bring delightful chaos to your Discord server!