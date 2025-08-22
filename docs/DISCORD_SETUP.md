# Discord Bot Setup Guide 🤖

This guide will walk you through creating a Discord application and bot from scratch, setting up permissions, and getting your bot token.

## Table of Contents

- [Creating a Discord Application](#creating-a-discord-application)
- [Bot Configuration](#bot-configuration)
- [Permissions Setup](#permissions-setup)
- [Inviting the Bot to Your Server](#inviting-the-bot-to-your-server)
- [Getting Your Bot Token](#getting-your-bot-token)
- [Server Requirements](#server-requirements)
- [Testing Your Setup](#testing-your-setup)

---

## Creating a Discord Application

1. **Go to Discord Developer Portal**
   - Visit https://discord.com/developers/applications
   - Log in with your Discord account

2. **Create New Application**
   - Click "New Application" button
   - Enter a name for your bot (e.g., "My Disruptor Bot")
   - Click "Create"

3. **Configure Application Details**
   - Add a description: "A fun bot that randomly plays soundboard sounds"
   - Upload an icon (optional)
   - Save changes

---

## Bot Configuration

1. **Navigate to Bot Settings**
   - In your application, click "Bot" in the left sidebar
   - Click "Add Bot" if not already present

2. **Configure Bot Settings**
   - **Username**: Choose a username for your bot
   - **Icon**: Upload a profile picture (optional)
   - **Public Bot**: ✅ **Disable** this if you only want to use the bot in your own servers
   - **Requires OAuth2 Code Grant**: ❌ **Keep disabled**

3. **Gateway Intents** (IMPORTANT)
   - ✅ **Server Members Intent** - Required for voice channel monitoring
   - ✅ **Message Content Intent** - Required for slash commands
   - ❌ **Presence Intent** - Not required

---

## Permissions Setup

Disruptor needs specific permissions to function properly:

### Required Permissions
- ✅ **View Channels** - To see voice channels
- ✅ **Connect** - To join voice channels  
- ✅ **Speak** - To play audio
- ✅ **Use Voice Activity** - For audio transmission
- ✅ **Use Slash Commands** - For bot control commands

### Recommended Permissions
- ✅ **Send Messages** - For command responses
- ✅ **Use External Emojis** - For rich responses
- ✅ **Read Message History** - For context

### Permission Value
The bot needs these permissions as a **minimum**:
```
Permissions Integer: 3148800
```

---

## Inviting the Bot to Your Server

1. **Generate Invite Link**
   - Go to "OAuth2" → "URL Generator" in your application
   - **Scopes**: Select `bot` and `applications.commands`
   - **Bot Permissions**: Select all the permissions listed above
   - Copy the generated URL

2. **Invite to Server**
   - Open the generated URL in a new tab
   - Select the server where you want to add the bot
   - Click "Authorize"
   - Complete any CAPTCHA if prompted

3. **Verify Bot Joined**
   - Check your server's member list
   - The bot should appear with an "APP" or "BOT" tag
   - The bot will be offline until you start the application

---

## Getting Your Bot Token

⚠️ **IMPORTANT**: Your bot token is like a password. Never share it publicly!

1. **Access Token**
   - Go to "Bot" section in your application
   - Under "Token" section, click "Copy"
   - Save this token securely - you'll need it for configuration

2. **Token Security**
   - ❌ Never commit tokens to Git repositories
   - ❌ Never share tokens in Discord messages or forums
   - ❌ Never screenshot tokens
   - ✅ Use environment variables to store tokens
   - ✅ Regenerate tokens if compromised

---

## Server Requirements

Your Discord server needs these features for Disruptor to work:

### Soundboard Setup
1. **Upload Sounds**
   - Go to Server Settings → Soundboard
   - Upload .mp3 or .ogg audio files (max 5.2MB each)
   - Name your sounds appropriately

2. **Sound Requirements**
   - At least 1 soundboard sound is required
   - Sounds should be short (1-10 seconds recommended)
   - Keep volume levels consistent

### Voice Channels
- At least 1 voice channel must exist
- Bot must have permissions in voice channels
- Voice channels should be accessible to bot

---

## Testing Your Setup

### 1. Basic Connection Test
```bash
# Set your bot token
export CONFIG_TOKEN="your_bot_token_here"

# Test basic connection (will fail but show connection attempt)
./output/bin/disruptor
```

**Expected Result**: Bot should connect to Discord and show online status

### 2. Verify Bot Permissions
- Check that bot appears online in your server
- Try using a slash command like `/play` (should appear in autocomplete)
- Verify bot can see voice channels

### 3. Test Soundboard Access
The bot will automatically discover soundboard sounds. Check logs for:
```
INF soundboard sounds discovered count=X guild_id=YOUR_GUILD_ID
```

---

## Common Issues

### Bot Won't Connect
- ✅ Check token is correct and copied fully
- ✅ Verify token hasn't been regenerated  
- ✅ Check internet connection

### Missing Slash Commands
- ✅ Ensure `applications.commands` scope was granted
- ✅ Wait up to 1 hour for global commands to sync
- ✅ Try kicking and re-inviting the bot

### Can't Join Voice Channels
- ✅ Verify "Connect" and "Speak" permissions
- ✅ Check voice channel user limits
- ✅ Ensure bot has access to the voice channel

### No Soundboard Sounds
- ✅ Upload at least one sound to server soundboard
- ✅ Verify sounds are not too large (5.2MB limit)
- ✅ Check bot has "Use Sounds from Other Servers" permission

---

## Next Steps

Once your Discord bot is set up:
1. Follow the [Installation Guide](INSTALLATION.md) to set up Disruptor
2. Check the [Configuration Guide](CONFIGURATION.md) for advanced settings
3. Read the [Quick Start Guide](QUICKSTART.md) for immediate setup

---

## Security Best Practices

- 🔐 Store bot token in environment variables
- 🔄 Regenerate token if compromised
- 👥 Limit bot to specific servers if needed
- 📝 Monitor bot activity through Discord audit logs
- 🛡️ Only grant necessary permissions

---

**Need Help?** Check the [Troubleshooting Guide](TROUBLESHOOTING.md) or open an issue on GitHub.