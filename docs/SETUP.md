# Setup Guide 🔧

Complete setup guide for Disruptor across different platforms and configurations. This guide covers detailed installation, configuration, and customization options.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Platform-Specific Installation](#platform-specific-installation)
- [Discord Bot Setup](#discord-bot-setup)
- [Configuration Options](#configuration-options)
- [Database Setup](#database-setup)
- [Advanced Configuration](#advanced-configuration)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)

## Prerequisites 📋

### System Requirements

- **RAM**: 256MB minimum, 512MB recommended
- **Storage**: 100MB for application, additional for database
- **Network**: Internet connection for Discord API access
- **OS**: Linux, macOS, Windows, or Docker

### Discord Requirements

- Discord server with **Administrator** permissions
- Server must support **Soundboard** feature (Nitro server or Community server)
- At least one soundboard sound uploaded

## Platform-Specific Installation 🖥️

### Docker (Recommended)

**Advantages**: Isolated environment, easy deployment, automatic dependency management

```bash
# Pull the latest image
docker pull ghcr.io/xanderd99/disruptor:latest

# Or build from source
git clone https://github.com/XanderD99/disruptor.git
cd disruptor
make docker-build
```

### Linux (Ubuntu/Debian)

```bash
# Install dependencies
sudo apt update
sudo apt install -y git golang-go libopus-dev pkg-config make

# Clone and build
git clone https://github.com/XanderD99/disruptor.git
cd disruptor
go mod download
make build build-migrate

# Install systemd service (optional)
sudo cp scripts/disruptor.service /etc/systemd/system/
sudo systemctl enable disruptor
```

### Linux (CentOS/RHEL/Fedora)

```bash
# Install dependencies
sudo dnf install -y git golang opus-devel pkgconf make

# Build process same as Ubuntu
git clone https://github.com/XanderD99/disruptor.git
cd disruptor
go mod download
make build build-migrate
```

### macOS

```bash
# Install dependencies with Homebrew
brew install go opus pkg-config make git

# Build process
git clone https://github.com/XanderD99/disruptor.git
cd disruptor
go mod download
make build build-migrate
```

### Windows

```powershell
# Install Go from https://golang.org/dl/
# Install Git from https://git-scm.com/

# Clone and build (in PowerShell)
git clone https://github.com/XanderD99/disruptor.git
cd disruptor
go mod download
go build -o disruptor.exe ./cmd/disruptor
go build -o migrate.exe ./cmd/migrate
```

**Note**: Opus library setup on Windows requires additional configuration. Docker is recommended for Windows users.

## Discord Bot Setup 🤖

### 1. Create Discord Application

1. Visit [Discord Developer Portal](https://discord.com/developers/applications)
2. Click **"New Application"**
3. Enter application name (e.g., "Disruptor")
4. Save the **Application ID** for later

### 2. Configure Bot Settings

1. Navigate to **"Bot"** section
2. Click **"Add Bot"**
3. **Copy the Token** (this is your `CONFIG_TOKEN`)
4. Configure bot settings:
   - ✅ **Public Bot**: Off (keep it private to your servers)
   - ✅ **Require OAuth2 Code Grant**: Off
   - ✅ **Server Members Intent**: On
   - ✅ **Presence Intent**: Off (not needed)
   - ✅ **Message Content Intent**: On

### 3. Set Bot Permissions

In **"OAuth2" → "URL Generator"**:

**Scopes**:

- ✅ `bot`
- ✅ `applications.commands`

**Bot Permissions**:

- ✅ **Connect** (join voice channels)
- ✅ **Speak** (play audio)
- ✅ **Use Slash Commands** (bot commands)
- ✅ **Manage Guild** (for `/weight` command)
- ✅ **View Channels** (see available channels)

### 4. Invite Bot to Server

1. Copy the generated OAuth2 URL
2. Open the URL in your browser
3. Select your Discord server
4. Click **"Authorize"**
5. Complete the CAPTCHA

### 5. Server Configuration

**Required Server Setup**:

1. **Soundboard Sounds**: Upload at least one sound
   - Server Settings → Soundboard → Upload Sound
   - Supported formats: MP3, OGG, WAV (max 5.2MB)

2. **Voice Channel Setup**:
   - Ensure bot has permissions in voice channels
   - Test with users in voice channels (bot needs other users present)

## Configuration Options ⚙️

### Environment Variables

All configuration uses the `CONFIG_` prefix:

#### **Required Configuration**

```bash
# Discord bot token (REQUIRED)
export CONFIG_TOKEN="your_discord_bot_token_here"
```

#### **Database Configuration**

```bash
# Database type (currently only SQLite supported)
export CONFIG_DATABASE_TYPE="sqlite"

# Database connection string
export CONFIG_DATABASE_DSN="file:./disruptor.db?cache=shared"

# Options:
# In-memory (dev):     file::memory:?cache=shared
# File (production):   file:./disruptor.db?cache=shared
# Custom path:         file:/path/to/database.db?cache=shared
```

#### **Logging Configuration**

```bash
# Log level: debug, info, warn, error
export CONFIG_LOGGING_LEVEL="info"

# Log format: json, text
export CONFIG_LOGGING_FORMAT="text"

# Enable colors in console output
export CONFIG_LOGGING_COLORS="true"
```

#### **Metrics Configuration**

```bash
# Enable Prometheus metrics server
export CONFIG_METRICS_ENABLED="true"

# Metrics server address
export CONFIG_METRICS_ADDRESS="0.0.0.0:8080"

# Metrics endpoint path
export CONFIG_METRICS_PATH="/metrics"
```

#### **Bot Behavior Configuration**

```bash
# Default disruption interval for new guilds
export CONFIG_DEFAULT_INTERVAL="1h"

# Default disruption chance for new guilds (0-100)
export CONFIG_DEFAULT_CHANCE="50"

# Maximum concurrent voice connections
export CONFIG_MAX_VOICE_CONNECTIONS="10"
```

### Configuration File

Create `.env` file for easier management:

```bash
# .env file
CONFIG_TOKEN=your_discord_bot_token_here
CONFIG_DATABASE_DSN=file:./disruptor.db?cache=shared
CONFIG_LOGGING_LEVEL=info
CONFIG_METRICS_ENABLED=true
CONFIG_METRICS_ADDRESS=0.0.0.0:8080
```

Load with:

```bash
# Load environment variables
source .env
# or
export $(cat .env | xargs)
```

## Database Setup 🗄️

### SQLite (Default)

**Development Setup**:

```bash
# In-memory database (data lost on restart)
export CONFIG_DATABASE_DSN="file::memory:?cache=shared"
```

**Production Setup**:

```bash
# Persistent file database
export CONFIG_DATABASE_DSN="file:./data/disruptor.db?cache=shared"

# Create data directory
mkdir -p data
chmod 755 data

# Run migrations
./output/bin/migrate up
```

**Database Files**:

- `disruptor.db`: Main database file
- `disruptor.db-shm`: Shared memory file
- `disruptor.db-wal`: Write-ahead log file

### Migration Management

```bash
# Build migration tool
make build-migrate

# Run all pending migrations
./output/bin/migrate up

# Check migration status
./output/bin/migrate status

# Reset database (careful!)
rm -f disruptor.db* && ./output/bin/migrate up
```

## Advanced Configuration 🚀

### Systemd Service (Linux)

Create `/etc/systemd/system/disruptor.service`:

```ini
[Unit]
Description=Disruptor Discord Bot
After=network.target
Wants=network.target

[Service]
Type=simple
User=disruptor
Group=disruptor
WorkingDirectory=/opt/disruptor
ExecStart=/opt/disruptor/disruptor
Environment=CONFIG_TOKEN=your_token_here
Environment=CONFIG_DATABASE_DSN=file:/opt/disruptor/data/disruptor.db?cache=shared
Environment=CONFIG_LOGGING_LEVEL=info
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable disruptor
sudo systemctl start disruptor
sudo systemctl status disruptor
```

### Reverse Proxy (Nginx)

For metrics endpoint:

```nginx
# /etc/nginx/sites-available/disruptor
server {
    listen 80;
    server_name disruptor-metrics.yourdomain.com;

    location /metrics {
        proxy_pass http://localhost:8080/metrics;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Optional: Basic auth for security
    auth_basic "Disruptor Metrics";
    auth_basic_user_file /etc/nginx/.htpasswd;
}
```

### Log Rotation

Create `/etc/logrotate.d/disruptor`:

```
/var/log/disruptor/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    create 644 disruptor disruptor
    postrotate
        systemctl reload disruptor
    endscript
}
```

## Verification ✅

### 1. Bot Status Check

```bash
# Check if bot is running
ps aux | grep disruptor

# Check logs
journalctl -u disruptor -f

# Docker logs
docker logs -f disruptor
```

### 2. Discord Integration Test

1. **Slash Commands**: Type `/` in your server - should see Disruptor commands
2. **Permissions**: Bot should appear online in member list
3. **Voice Access**: Bot should be able to join voice channels

### 3. Functionality Test

```bash
# Test commands
/play                    # Should play a random sound
/interval 5m            # Should set 5-minute interval
/chance 100             # Should set 100% chance
/next                   # Should show next scheduled disruption
/weight #channel 75     # Should set channel weight
```

### 4. Metrics Check

```bash
# Check metrics endpoint
curl http://localhost:8080/metrics

# Should return Prometheus metrics format
# Look for disruptor_* metrics
```

## Troubleshooting 🔧

### Common Issues

#### Bot Won't Start

```bash
# Check token validity
curl -H "Authorization: Bot $CONFIG_TOKEN" \
  https://discord.com/api/v10/users/@me

# Check environment variables
env | grep CONFIG_

# Check file permissions
ls -la disruptor
ls -la data/
```

#### Database Issues

```bash
# Check database file
ls -la *.db*

# Reset database
rm -f disruptor.db* && ./output/bin/migrate up

# Check database content
sqlite3 disruptor.db ".tables"
sqlite3 disruptor.db "SELECT * FROM guilds;"
```

#### Discord Permission Issues

1. **Re-invite bot** with correct permissions
2. **Check role hierarchy** - bot role should be above member roles
3. **Verify channel permissions** - bot needs Connect/Speak in voice channels

#### No Slash Commands

- **Wait up to 1 hour** for Discord to sync global commands
- **Re-invite bot** with `applications.commands` scope
- **Check bot permissions** in server settings

#### Bot Leaves Voice Channel Immediately

- **Need other users** in voice channel (bot won't stay alone)
- **Check soundboard** has uploaded sounds
- **Verify voice permissions** (Connect + Speak)

### Debug Mode

```bash
# Enable detailed logging
export CONFIG_LOGGING_LEVEL="debug"

# Monitor specific components
grep "scheduler" /var/log/disruptor/disruptor.log
grep "voice" /var/log/disruptor/disruptor.log
```

### Getting Help

- 📖 **Documentation**: [Full docs](../README.md)
- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/XanderD99/disruptor/issues)
- 💬 **Questions**: [GitHub Discussions](https://github.com/XanderD99/disruptor/discussions)
- 🚀 **Feature Requests**: [GitHub Issues](https://github.com/XanderD99/disruptor/issues)

---

**🎉 Setup Complete!** Your Disruptor bot should now be ready to bring delightful chaos to your Discord server!
