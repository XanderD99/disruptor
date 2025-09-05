# Quick Start Guide 🚀

Get Disruptor running in 10 minutes! This guide gets you from zero to a working Discord bot that randomly joins voice channels and plays soundboard sounds.

## Prerequisites 📋

- [Discord Bot Token](#step-1-create-discord-bot) (free from Discord Developer Portal)
- **Option A**: [Docker](https://docker.com/get-started) (easiest)
- **Option B**: [Go 1.25+](https://golang.org/dl/) + [Git](https://git-scm.com/downloads) (for building from source)

## Step 1: Create Discord Bot 🤖

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Click **"New Application"** → Enter name **"Disruptor"**
3. Navigate to **"Bot"** section → Click **"Add Bot"**
4. Copy the **Token** (keep this secret!)
5. Under **"Privileged Gateway Intents"** enable:
   - ✅ Server Members Intent
   - ✅ Message Content Intent

## Step 2: Invite Bot to Server 🏠

1. In **"OAuth2" → "URL Generator"** select:
   - ✅ **bot**
   - ✅ **applications.commands**
2. Under **Bot Permissions** select:
   - ✅ Connect
   - ✅ Speak
   - ✅ Use Slash Commands
   - ✅ Manage Guild (for `/weight` command)
3. Copy the generated URL and open it
4. Select your server and authorize

## Step 3A: Run with Docker (Recommended) 🐳

**Single command setup:**

```bash
# Replace YOUR_BOT_TOKEN with your actual token
docker run -d --name disruptor \
  -e CONFIG_TOKEN="YOUR_BOT_TOKEN" \
  -e CONFIG_DATABASE_DSN="file:/data/disruptor.db?cache=shared" \
  -v $(pwd)/data:/data \
  --restart unless-stopped \
  ghcr.io/xanderd99/disruptor:latest
```

**Or with docker-compose:**

```bash
# Create docker-compose.yml
cat > docker-compose.yml << 'EOF'
services:
  disruptor:
    image: ghcr.io/xanderd99/disruptor:latest
    environment:
      - CONFIG_TOKEN=${DISCORD_TOKEN}
      - CONFIG_DATABASE_DSN=file:/data/disruptor.db?cache=shared
      - CONFIG_LOGGING_LEVEL=info
    volumes:
      - ./data:/data
    restart: unless-stopped
EOF

# Create .env file with your token
echo "DISCORD_TOKEN=YOUR_BOT_TOKEN" > .env

# Start the bot
docker-compose up -d
```

## Step 3B: Build from Source 🔨

```bash
# Clone and build
git clone https://github.com/XanderD99/disruptor.git
cd disruptor
go mod download
make build

# Run with your token
export CONFIG_TOKEN="YOUR_BOT_TOKEN"
export CONFIG_DATABASE_DSN="file:./disruptor.db?cache=shared"
./output/bin/disruptor
```

## Step 4: Add Soundboard Sounds 🎵

1. In your Discord server, go to any channel
2. Right-click → **Server Settings** → **Soundboard**
3. Click **"Upload Sound"** and add some audio files
4. Give them fun names!

## Step 5: Test the Bot 🎮

1. Join a voice channel with at least one other person
2. Use slash commands:

   ```ansii
   /play                    # Play a sound immediately
   /interval 30m           # Set disruption interval to 30 minutes
   /chance 75              # Set 75% disruption chance
   /weight #channel 80     # Make this channel more likely to be selected
   /next                   # See when next disruption is scheduled
   ```

3. Wait for the magic! The bot will randomly join and play sounds

## Default Settings ⚙️

- **Interval**: 1 hour (random disruptions every ~1 hour)
- **Chance**: 50% (50% probability of disruption per interval)
- **Channel Weight**: 50 (balanced selection for all channels)
- **Database**: In-memory (resets on restart)

## Quick Troubleshooting 🔧

| Problem | Solution |
|---------|----------|
| Bot won't start | Double-check your `CONFIG_TOKEN` |
| No slash commands | Wait up to 1 hour for Discord to sync, or re-invite bot |
| Bot won't join voice | Ensure bot has **Connect** and **Speak** permissions |
| No sounds play | Upload sounds to your server's Discord soundboard |
| Bot leaves immediately | Need at least 1 other person in the voice channel |

## What's Next? 🎯

- **Customize disruptions**: Adjust `/interval` and `/chance` per your server's vibe
- **Set channel weights**: Use `/weight` to make certain channels more/less likely
- **Monitor activity**: Check Docker logs with `docker logs disruptor`
- **Advanced setup**: See [Setup Guide](SETUP.md) for detailed configuration
- **Production deployment**: See [Deployment Guide](DEPLOYMENT.md)

---

**🎉 Congratulations!** Your Discord server is now delightfully disrupted. Enjoy the chaos responsibly!

**Need help?** Check out the [full documentation](README.md) or [create an issue](https://github.com/XanderD99/disruptor/issues).
