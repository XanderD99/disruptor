# Configuration Guide ⚙️

This guide covers all configuration options for Disruptor, with practical examples and best practices.

## Table of Contents

- [Configuration Methods](#configuration-methods)
- [Essential Configuration](#essential-configuration)
- [Environment Variables](#environment-variables)
- [Database Configuration](#database-configuration)
- [Logging Configuration](#logging-configuration)
- [Advanced Configuration](#advanced-configuration)
- [Configuration Examples](#configuration-examples)
- [Best Practices](#best-practices)

---

## Configuration Methods

Disruptor uses **environment variables** for all configuration. You can set these through:

1. **Shell Environment**
   ```bash
   export CONFIG_TOKEN="your_token"
   ./disruptor
   ```

2. **Environment File**
   ```bash
   # Create .env file
   echo "CONFIG_TOKEN=your_token" > .env
   
   # Load and run
   set -a; source .env; set +a
   ./disruptor
   ```

3. **Docker Environment**
   ```bash
   docker run -e CONFIG_TOKEN="your_token" disruptor
   ```

4. **Systemd Service**
   ```ini
   [Service]
   Environment=CONFIG_TOKEN=your_token
   EnvironmentFile=/opt/disruptor/.env
   ```

---

## Essential Configuration

### Required Settings

#### Discord Bot Token
```bash
CONFIG_TOKEN="your_discord_bot_token_here"
```
- **Required**: Yes
- **Description**: Your Discord bot token from Discord Developer Portal
- **Security**: Never commit this to version control!

### Recommended Settings

#### Database Storage
```bash
# File-based database (recommended for production)
CONFIG_DATABASE_DSN="file:./disruptor.db?cache=shared"

# Custom path
CONFIG_DATABASE_DSN="file:/opt/disruptor/data/bot.db?cache=shared"
```

#### Logging Level
```bash
# Production
CONFIG_LOGGING_LEVEL="info"

# Development
CONFIG_LOGGING_LEVEL="debug"

# Quiet
CONFIG_LOGGING_LEVEL="warn"
```

---

## Environment Variables

### Discord Configuration

#### Bot Token (Required)
```bash
CONFIG_TOKEN="your_discord_bot_token"
```

#### Sharding (Advanced)
```bash
# Single shard (default)
CONFIG_SHARDING_IDS="0"
CONFIG_SHARDING_COUNT="1"
CONFIG_SHARDING_AUTOSCALING="false"

# Multiple shards (large bots)
CONFIG_SHARDING_IDS="0,1,2,3"
CONFIG_SHARDING_COUNT="4"
CONFIG_SHARDING_AUTOSCALING="true"
```

### Database Configuration

#### SQLite (Default)
```bash
CONFIG_DATABASE_TYPE="sqlite"

# In-memory (development)
CONFIG_DATABASE_DSN="file::memory:?cache=shared"

# File-based (production)
CONFIG_DATABASE_DSN="file:./disruptor.db?cache=shared"

# Custom location
CONFIG_DATABASE_DSN="file:/var/lib/disruptor/bot.db?cache=shared"
```

### Logging Configuration

#### Basic Logging
```bash
CONFIG_LOGGING_LEVEL="info"           # debug, info, warn, error
CONFIG_LOGGING_PRETTY="true"          # Human-readable logs
CONFIG_LOGGING_COLORS="true"          # Colored output
CONFIG_LOGGING_SOURCE="false"         # Include source file info
```

#### Discord Webhook Logging
```bash
CONFIG_LOGGING_DISCORD_WEBHOOK="https://discord.com/api/webhooks/..."
CONFIG_LOGGING_DISCORD_MIN_LEVEL="warn"  # Only send warnings and errors
CONFIG_LOGGING_DISCORD_SYNC="false"      # Don't block on webhook delivery
```

### Metrics Configuration

#### Prometheus Metrics
```bash
CONFIG_METRICS_PORT="9090"             # Metrics server port
CONFIG_METRICS_SHUTDOWN_DURATION="15s" # Graceful shutdown timeout
```

### Lavalink Configuration (Optional)

```bash
CONFIG_LAVALINK_NODENAME="disruptor"
CONFIG_LAVALINK_NODEADDRESS="localhost:2333"
CONFIG_LAVALINK_NODEPASSWORD="youshallnotpass"
CONFIG_LAVALINK_NODESECURE="false"
```

---

## Database Configuration

### SQLite Options

#### Development Setup
```bash
# In-memory database (data lost on restart)
CONFIG_DATABASE_DSN="file::memory:?cache=shared"
```
**Use for**: Testing, development, temporary instances

#### Production Setup
```bash
# Persistent file database
CONFIG_DATABASE_DSN="file:./disruptor.db?cache=shared"
```
**Use for**: Production, persistent bot configuration

#### Advanced SQLite Options
```bash
# With additional SQLite parameters
CONFIG_DATABASE_DSN="file:./disruptor.db?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000"
```

**Parameters explained**:
- `cache=shared`: Share cache between connections
- `mode=rwc`: Read/write/create mode
- `_journal_mode=WAL`: Write-Ahead Logging for better performance
- `_busy_timeout=5000`: Wait 5 seconds for lock resolution

### Database Location Examples

```bash
# Current directory
CONFIG_DATABASE_DSN="file:./disruptor.db?cache=shared"

# Absolute path
CONFIG_DATABASE_DSN="file:/opt/disruptor/data/bot.db?cache=shared"

# User home directory
CONFIG_DATABASE_DSN="file:$HOME/.disruptor/bot.db?cache=shared"

# System directory (requires permissions)
CONFIG_DATABASE_DSN="file:/var/lib/disruptor/bot.db?cache=shared"
```

---

## Logging Configuration

### Log Levels

```bash
# Development - All messages
CONFIG_LOGGING_LEVEL="debug"

# Production - Important messages only
CONFIG_LOGGING_LEVEL="info"

# Production - Warnings and errors only
CONFIG_LOGGING_LEVEL="warn"

# Production - Errors only
CONFIG_LOGGING_LEVEL="error"
```

### Log Format Options

#### Human-Readable (Development)
```bash
CONFIG_LOGGING_PRETTY="true"
CONFIG_LOGGING_COLORS="true"
CONFIG_LOGGING_SOURCE="true"
```

**Output Example**:
```
Aug 22 18:51:11.270 INF process manager initialized max_groups=0
```

#### JSON Format (Production)
```bash
CONFIG_LOGGING_PRETTY="false"
CONFIG_LOGGING_COLORS="false"
CONFIG_LOGGING_SOURCE="false"
```

**Output Example**:
```json
{"time":"2024-08-22T18:51:11.270Z","level":"INFO","msg":"process manager initialized","max_groups":0}
```

### Discord Webhook Logging

Send logs to Discord channel for monitoring:

```bash
CONFIG_LOGGING_DISCORD_WEBHOOK="https://discord.com/api/webhooks/123456789/abcdefg"
CONFIG_LOGGING_DISCORD_MIN_LEVEL="warn"  # 4 = warn level
CONFIG_LOGGING_DISCORD_SYNC="false"
```

**Log Levels for Discord**:
- `1` = debug
- `2` = info  
- `3` = warn
- `4` = error

---

## Advanced Configuration

### Performance Tuning

#### Memory Optimization
```bash
# Go runtime settings
export GOGC=100              # Garbage collection target
export GOMAXPROCS=2          # CPU cores to use
```

#### Database Performance
```bash
CONFIG_DATABASE_DSN="file:./disruptor.db?cache=shared&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=-64000"
```

### Security Hardening

#### File Permissions
```bash
# Secure database file
chmod 600 disruptor.db

# Secure config file
chmod 600 .env
```

#### User Isolation
```bash
# Run as dedicated user
sudo useradd -r -s /bin/false disruptor
sudo chown disruptor:disruptor /opt/disruptor/
```

---

## Configuration Examples

### Development Environment

```bash
# .env.development
CONFIG_TOKEN="your_dev_bot_token"
CONFIG_DATABASE_DSN="file::memory:?cache=shared"
CONFIG_LOGGING_LEVEL="debug"
CONFIG_LOGGING_PRETTY="true"
CONFIG_LOGGING_COLORS="true"
CONFIG_LOGGING_SOURCE="true"
CONFIG_METRICS_PORT="9090"
```

### Production Environment

```bash
# .env.production
CONFIG_TOKEN="your_production_bot_token"
CONFIG_DATABASE_DSN="file:/var/lib/disruptor/bot.db?cache=shared&_journal_mode=WAL"
CONFIG_LOGGING_LEVEL="info"
CONFIG_LOGGING_PRETTY="false"
CONFIG_LOGGING_COLORS="false"
CONFIG_LOGGING_SOURCE="false"
CONFIG_LOGGING_DISCORD_WEBHOOK="https://discord.com/api/webhooks/..."
CONFIG_LOGGING_DISCORD_MIN_LEVEL="warn"
CONFIG_METRICS_PORT="9090"
```

### High-Availability Setup

```bash
# .env.ha
CONFIG_TOKEN="your_bot_token"
CONFIG_DATABASE_DSN="file:/shared/storage/disruptor.db?cache=shared&_journal_mode=WAL"
CONFIG_LOGGING_LEVEL="info"
CONFIG_LOGGING_DISCORD_WEBHOOK="https://discord.com/api/webhooks/..."
CONFIG_METRICS_PORT="9090"
CONFIG_SHARDING_AUTOSCALING="true"
```

### Docker Compose Configuration

```yaml
# docker-compose.yml
version: '3.8'
services:
  disruptor:
    image: disruptor:latest
    environment:
      - CONFIG_TOKEN=${DISCORD_TOKEN}
      - CONFIG_DATABASE_DSN=file:/data/disruptor.db?cache=shared
      - CONFIG_LOGGING_LEVEL=info
      - CONFIG_LOGGING_PRETTY=false
      - CONFIG_METRICS_PORT=9090
    volumes:
      - ./data:/data
    ports:
      - "9090:9090"
    restart: unless-stopped
```

### Kubernetes Configuration

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: disruptor-config
data:
  CONFIG_LOGGING_LEVEL: "info"
  CONFIG_LOGGING_PRETTY: "false"
  CONFIG_DATABASE_DSN: "file:/data/disruptor.db?cache=shared"
  CONFIG_METRICS_PORT: "9090"

---
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: disruptor-secrets
type: Opaque
stringData:
  CONFIG_TOKEN: "your_discord_bot_token"
```

---

## Best Practices

### Security

1. **Never Hardcode Tokens**
   ```bash
   # ❌ Don't do this
   CONFIG_TOKEN="ODc2..." ./disruptor
   
   # ✅ Do this
   source .env && ./disruptor
   ```

2. **Secure File Permissions**
   ```bash
   chmod 600 .env
   chmod 600 disruptor.db
   ```

3. **Use Dedicated User**
   ```bash
   sudo useradd -r disruptor
   sudo -u disruptor ./disruptor
   ```

### Performance

1. **Use File Database in Production**
   ```bash
   # ❌ Don't use in-memory for production
   CONFIG_DATABASE_DSN="file::memory:?cache=shared"
   
   # ✅ Use persistent storage
   CONFIG_DATABASE_DSN="file:./disruptor.db?cache=shared"
   ```

2. **Optimize SQLite Settings**
   ```bash
   CONFIG_DATABASE_DSN="file:./disruptor.db?cache=shared&_journal_mode=WAL&_synchronous=NORMAL"
   ```

3. **Appropriate Log Levels**
   ```bash
   # Development
   CONFIG_LOGGING_LEVEL="debug"
   
   # Production
   CONFIG_LOGGING_LEVEL="info"
   ```

### Monitoring

1. **Enable Metrics**
   ```bash
   CONFIG_METRICS_PORT="9090"
   # Access metrics at http://localhost:9090/metrics
   ```

2. **Discord Webhook Monitoring**
   ```bash
   CONFIG_LOGGING_DISCORD_WEBHOOK="https://discord.com/api/webhooks/..."
   CONFIG_LOGGING_DISCORD_MIN_LEVEL="warn"
   ```

3. **Structured Logging for Production**
   ```bash
   CONFIG_LOGGING_PRETTY="false"  # JSON format for log aggregation
   ```

### Maintenance

1. **Database Backups**
   ```bash
   # Regular backups
   cp disruptor.db "disruptor.db.backup.$(date +%Y%m%d_%H%M%S)"
   ```

2. **Configuration Validation**
   ```bash
   # Test configuration without running
   ./disruptor --validate-config
   ```

3. **Gradual Rollouts**
   ```bash
   # Test with single shard first
   CONFIG_SHARDING_COUNT="1"
   ```

---

## Configuration Validation

### Check Current Configuration
```bash
# See all environment variables
env | grep CONFIG_

# Validate configuration (if supported)
./disruptor --validate
```

### Common Issues

#### Token Issues
```bash
# Test token validity
export CONFIG_TOKEN="your_token"
timeout 10s ./disruptor
# Should connect to Discord without errors
```

#### Database Issues
```bash
# Test database connection
CONFIG_DATABASE_DSN="file:./test.db?cache=shared" ./disruptor --test-db
```

#### Permission Issues
```bash
# Check file permissions
ls -la disruptor.db
ls -la .env
```

---

## Getting Help

- 📖 **Reference**: See [Environment Variables](ENVIRONMENT.md) for complete list
- 🚀 **Quick Start**: Try [Quick Start Guide](QUICKSTART.md) for immediate setup
- 🐛 **Issues**: Check [Troubleshooting Guide](TROUBLESHOOTING.md)
- 💬 **Support**: Open GitHub issue for configuration help

---

**Configuration complete!** 🎉 Your Disruptor bot is ready to run.