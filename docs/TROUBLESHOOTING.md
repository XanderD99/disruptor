# Troubleshooting Guide 🛠️

This guide helps you diagnose and fix common issues with Disruptor.

## Table of Contents

- [Quick Diagnostics](#quick-diagnostics)
- [Bot Connection Issues](#bot-connection-issues)
- [Audio/Voice Issues](#audiovoice-issues)
- [Database Issues](#database-issues)
- [Build and Installation Issues](#build-and-installation-issues)
- [Performance Issues](#performance-issues)
- [Configuration Issues](#configuration-issues)
- [Discord API Issues](#discord-api-issues)
- [Getting Help](#getting-help)

---

## Quick Diagnostics

### Health Check Script
```bash
#!/bin/bash
# Save as health-check.sh

echo "🔍 Disruptor Health Check"
echo "========================="

# Check if binary exists
if [ -f "./output/bin/disruptor" ]; then
    echo "✅ Binary found"
else
    echo "❌ Binary not found - run 'make build'"
    exit 1
fi

# Check Go version
echo "📋 Go version: $(go version)"

# Check if CONFIG_TOKEN is set
if [ -n "$CONFIG_TOKEN" ]; then
    echo "✅ CONFIG_TOKEN is set"
else
    echo "❌ CONFIG_TOKEN not set"
fi

# Check database file
if [ -n "$CONFIG_DATABASE_DSN" ] && [[ "$CONFIG_DATABASE_DSN" == *"file:"* ]]; then
    DB_PATH=$(echo "$CONFIG_DATABASE_DSN" | sed 's/.*file:\([^?]*\).*/\1/')
    if [ -f "$DB_PATH" ]; then
        echo "✅ Database file exists: $DB_PATH"
    else
        echo "⚠️  Database file will be created: $DB_PATH"
    fi
fi

# Test basic connectivity
echo "🔌 Testing basic startup..."
timeout 10s ./output/bin/disruptor 2>&1 | head -n 5

echo "✅ Health check complete"
```

### Log Analysis Script
```bash
#!/bin/bash
# Save as analyze-logs.sh

echo "📋 Analyzing Disruptor Logs"
echo "==========================="

# Systemd logs
if systemctl is-active --quiet disruptor; then
    echo "📊 Service Status: ✅ Running"
    echo "📊 Last 10 log entries:"
    sudo journalctl -u disruptor -n 10 --no-pager
    
    echo -e "\n🔍 Error Summary:"
    sudo journalctl -u disruptor --since "1 hour ago" | grep -i error | tail -5
    
    echo -e "\n⚠️  Warning Summary:"
    sudo journalctl -u disruptor --since "1 hour ago" | grep -i warn | tail -5
else
    echo "📊 Service Status: ❌ Not running"
fi

# Docker logs
if docker ps | grep -q disruptor; then
    echo -e "\n🐳 Docker Status: ✅ Running"
    echo "📊 Last 10 log entries:"
    docker logs disruptor --tail 10
    
    echo -e "\n🔍 Error Summary:"
    docker logs disruptor --since 1h | grep -i error | tail -5
fi
```

---

## Bot Connection Issues

### Issue: "Required environment variable CONFIG_TOKEN is not set"

**Symptoms:**
```
Error loading configuration: env: required environment variable "CONFIG_TOKEN" is not set
```

**Solutions:**
```bash
# Method 1: Export environment variable
export CONFIG_TOKEN="your_discord_bot_token"
./output/bin/disruptor

# Method 2: Create .env file
echo "CONFIG_TOKEN=your_discord_bot_token" > .env
source .env
./output/bin/disruptor

# Method 3: Inline
CONFIG_TOKEN="your_discord_bot_token" ./output/bin/disruptor
```

**Verification:**
```bash
echo $CONFIG_TOKEN  # Should show your token (first few characters)
```

### Issue: "Illegal base64 data at input byte"

**Symptoms:**
```
error creating Discord bot: error while getting application id from token: illegal base64 data at input byte 4
```

**Causes:**
- Invalid bot token format
- Token copied incorrectly
- Token has been regenerated

**Solutions:**
1. **Verify Token Format**
   ```bash
   # Token should be long and contain dots
   echo $CONFIG_TOKEN | wc -c  # Should be around 70+ characters
   ```

2. **Re-copy Token from Discord**
   - Go to Discord Developer Portal
   - Select your application
   - Go to "Bot" section
   - Click "Copy" under Token section
   - Paste carefully without extra spaces

3. **Check for Hidden Characters**
   ```bash
   # Check for hidden characters
   echo "$CONFIG_TOKEN" | hexdump -C | head
   ```

### Issue: Bot Goes Offline Immediately

**Symptoms:**
- Bot shows online briefly then goes offline
- Logs show connection established then disconnection

**Common Causes & Solutions:**

1. **Token Issues**
   ```bash
   # Test token validity
   curl -H "Authorization: Bot $CONFIG_TOKEN" \
        https://discord.com/api/v10/users/@me
   ```

2. **Network Issues**
   ```bash
   # Check connectivity to Discord
   ping discord.com
   nslookup discord.com
   ```

3. **Rate Limiting**
   ```
   # Look for rate limit messages in logs
   grep -i "rate limit" logs
   ```

4. **Permissions Issues**
   - Verify bot has required permissions in Discord
   - Check if bot was kicked from servers

---

## Audio/Voice Issues

### Issue: Bot Can't Join Voice Channels

**Symptoms:**
- Bot doesn't join voice channel when command is used
- "Permission denied" errors in logs

**Solutions:**

1. **Check Bot Permissions**
   ```
   Required permissions:
   ✅ View Channels
   ✅ Connect
   ✅ Speak  
   ✅ Use Voice Activity
   ```

2. **Verify Voice Channel Access**
   - Bot must have access to specific voice channel
   - Check channel-specific permission overrides
   - Ensure channel isn't full (user limit)

3. **Check Audio Dependencies**
   ```bash
   # Linux: Check for opus library
   pkg-config --modversion opus
   
   # Should show version like: 1.3.1
   ```

### Issue: No Sound Playing

**Symptoms:**
- Bot joins voice channel but no audio plays
- Soundboard commands execute but silent

**Solutions:**

1. **Check Soundboard Setup**
   ```
   Requirements:
   - At least 1 soundboard sound uploaded to server
   - Sounds must be under 5.2MB
   - Supported formats: MP3, OGG
   ```

2. **Verify Soundboard Permissions**
   - Bot needs "Use Soundboard" permission
   - Check server settings for soundboard restrictions

3. **Check Bot Voice State**
   ```bash
   # Look for voice state logs
   grep -i "voice" logs | grep -i "state"
   ```

### Issue: Audio Cuts Out or Poor Quality

**Symptoms:**
- Audio plays but cuts out frequently
- Poor audio quality or distortion

**Solutions:**

1. **Check Network Bandwidth**
   ```bash
   # Test network speed
   speedtest-cli
   
   # Check for packet loss
   ping -c 10 discord.com
   ```

2. **Reduce Audio Quality** (if needed)
   ```bash
   # Add to configuration (if supported)
   CONFIG_AUDIO_BITRATE=64000
   ```

3. **Check System Resources**
   ```bash
   # Monitor CPU/memory usage
   top -p $(pgrep disruptor)
   ```

---

## Database Issues

### Issue: Database File Permissions

**Symptoms:**
```
error opening database: unable to open database file
```

**Solutions:**
```bash
# Check file permissions
ls -la disruptor.db

# Fix ownership (systemd deployment)
sudo chown disruptor:disruptor /opt/disruptor/data/disruptor.db

# Fix permissions
chmod 644 disruptor.db

# Check directory permissions
ls -la $(dirname "disruptor.db")
```

### Issue: Database Corruption

**Symptoms:**
```
database disk image is malformed
```

**Solutions:**

1. **Try SQLite Recovery**
   ```bash
   # Backup corrupted database
   cp disruptor.db disruptor.db.corrupted
   
   # Try to repair
   sqlite3 disruptor.db ".recover" | sqlite3 disruptor_recovered.db
   
   # Replace if successful
   mv disruptor_recovered.db disruptor.db
   ```

2. **Restore from Backup**
   ```bash
   # If you have backups
   cp disruptor.db.backup.20240101 disruptor.db
   ```

3. **Start Fresh**
   ```bash
   # Remove corrupted database (will lose data)
   rm disruptor.db
   # Bot will create new database on startup
   ```

### Issue: Database Lock

**Symptoms:**
```
database is locked
```

**Solutions:**
```bash
# Check for other processes using the database
lsof disruptor.db

# Kill processes if needed
sudo pkill disruptor

# Check for .wal or .shm files
ls -la disruptor.db*

# Remove lock files (ONLY if bot is stopped)
rm disruptor.db-wal disruptor.db-shm
```

---

## Build and Installation Issues

### Issue: "make: command not found"

**Solutions:**
```bash
# Ubuntu/Debian
sudo apt install make

# macOS
xcode-select --install
# or
brew install make

# Windows
choco install make
```

### Issue: "opus.h: No such file or directory"

**Solutions:**
```bash
# Ubuntu/Debian
sudo apt install libopus-dev pkg-config

# CentOS/RHEL
sudo yum install opus-devel pkgconfig

# macOS
brew install opus pkg-config

# Verify installation
pkg-config --modversion opus
```

### Issue: CGO Build Failures

**Symptoms:**
```
CGO disabled
exec: "gcc": executable file not found
```

**Solutions:**
```bash
# Install GCC
# Ubuntu/Debian
sudo apt install build-essential

# CentOS/RHEL
sudo yum groupinstall "Development Tools"

# macOS
xcode-select --install

# Enable CGO
export CGO_ENABLED=1

# Verify
go env CGO_ENABLED
```

### Issue: Go Version Too Old

**Symptoms:**
```
go: module requires go 1.24
```

**Solutions:**
```bash
# Check current version
go version

# Install latest Go (Linux)
wget https://go.dev/dl/go1.24.4.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# macOS
brew upgrade go

# Windows: Download installer from golang.org
```

---

## Performance Issues

### Issue: High Memory Usage

**Symptoms:**
- Bot using excessive RAM
- System running out of memory

**Diagnostics:**
```bash
# Check memory usage
ps aux | grep disruptor
top -p $(pgrep disruptor)

# Check Go memory stats (if exposed in metrics)
curl http://localhost:9090/metrics | grep go_memstats
```

**Solutions:**
```bash
# Tune Go garbage collector
export GOGC=50  # More aggressive GC

# Limit memory (systemd)
[Service]
MemoryMax=512M

# Container limits
docker run --memory=512m disruptor
```

### Issue: High CPU Usage

**Symptoms:**
- Bot consuming excessive CPU
- System becoming unresponsive

**Diagnostics:**
```bash
# Check CPU usage
top -p $(pgrep disruptor)

# Profile if available
go tool pprof http://localhost:9090/debug/pprof/profile
```

**Solutions:**
```bash
# Limit CPU cores
export GOMAXPROCS=1

# Container limits
docker run --cpus="0.5" disruptor

# Check for infinite loops in logs
grep -i "loop\|panic\|fatal" logs
```

### Issue: Slow Response Times

**Symptoms:**
- Slash commands take long to respond
- Bot seems sluggish

**Solutions:**
```bash
# Check database performance
sqlite3 disruptor.db "PRAGMA optimize;"
sqlite3 disruptor.db "VACUUM;"

# Enable WAL mode for better concurrency
CONFIG_DATABASE_DSN="file:./disruptor.db?cache=shared&_journal_mode=WAL"

# Check network latency to Discord
ping discord.com
```

---

## Configuration Issues

### Issue: Environment Variables Not Loading

**Symptoms:**
- Configuration seems ignored
- Default values used instead of set values

**Diagnostics:**
```bash
# Check environment
env | grep CONFIG_

# Test variable access
echo $CONFIG_LOGGING_LEVEL
```

**Solutions:**
```bash
# Ensure variables are exported
export CONFIG_LOGGING_LEVEL=debug

# Source .env file properly
set -a; source .env; set +a

# Check for typos in variable names
# Correct: CONFIG_TOKEN
# Wrong: CONFIG_TOKEn, CONF_TOKEN
```

### Issue: Invalid Configuration Values

**Symptoms:**
```
invalid value for CONFIG_LOGGING_LEVEL: "Debug" (expected: debug, info, warn, error)
```

**Solutions:**
```bash
# Use correct case
CONFIG_LOGGING_LEVEL=debug  # not Debug or DEBUG

# Check valid values in documentation
CONFIG_LOGGING_LEVEL=debug   # ✅
CONFIG_LOGGING_LEVEL=info    # ✅
CONFIG_LOGGING_LEVEL=warn    # ✅
CONFIG_LOGGING_LEVEL=error   # ✅
```

---

## Discord API Issues

### Issue: Rate Limiting

**Symptoms:**
```
429 Too Many Requests
rate limit exceeded
```

**Solutions:**
- Bot automatically handles rate limits
- Reduce command frequency if manual testing
- Check for multiple bot instances using same token

### Issue: Permission Denied

**Symptoms:**
```
403 Forbidden
Missing Access
```

**Solutions:**
1. **Re-invite Bot with Correct Permissions**
   - Generate new invite URL with required permissions
   - Remove and re-add bot to server

2. **Check Channel-Specific Permissions**
   - Verify bot role has permissions in specific channels
   - Check for permission overrides

---

## Getting Help

### Collect Diagnostic Information

Before asking for help, collect:

```bash
# System information
uname -a
go version
./output/bin/disruptor --version

# Configuration (hide sensitive data)
env | grep CONFIG_ | sed 's/TOKEN=.*/TOKEN=***/'

# Recent logs
sudo journalctl -u disruptor -n 50 --no-pager

# Service status
sudo systemctl status disruptor

# Resource usage
ps aux | grep disruptor
df -h
free -h
```

### Log Levels for Debugging

```bash
# Enable debug logging
CONFIG_LOGGING_LEVEL=debug
CONFIG_LOGGING_SOURCE=true

# Run and collect logs
./output/bin/disruptor 2>&1 | tee debug.log
```

### Support Channels

1. **GitHub Issues**: https://github.com/XanderD99/disruptor/issues
   - Bug reports
   - Feature requests
   - Include diagnostic information

2. **GitHub Discussions**: Use for:
   - Setup questions
   - Configuration help
   - General troubleshooting

3. **Documentation**: Check other guides:
   - [Installation Guide](INSTALLATION.md)
   - [Configuration Guide](CONFIGURATION.md)
   - [Deployment Guide](DEPLOYMENT.md)

### Bug Report Template

```markdown
**Environment:**
- OS: [Ubuntu 20.04 / macOS / Windows 10]
- Go Version: [go version output]
- Deployment: [binary / docker / kubernetes]

**Configuration:**
```bash
# Include relevant CONFIG_ variables (hide sensitive data)
CONFIG_LOGGING_LEVEL=debug
CONFIG_DATABASE_DSN=file:./disruptor.db?cache=shared
```

**Issue Description:**
[Detailed description of the problem]

**Steps to Reproduce:**
1. [First step]
2. [Second step]
3. [Issue occurs]

**Expected Behavior:**
[What should happen]

**Actual Behavior:**
[What actually happens]

**Logs:**
```
[Include relevant log entries]
```

**Additional Context:**
[Any other relevant information]
```

---

## Common Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| `CONFIG_TOKEN is not set` | Missing bot token | Set `CONFIG_TOKEN` environment variable |
| `illegal base64 data` | Invalid token format | Re-copy token from Discord Developer Portal |
| `database is locked` | Multiple processes or crash | Stop all instances, remove lock files |
| `permission denied` | File permissions | Fix file ownership and permissions |
| `opus.h: No such file` | Missing audio library | Install libopus-dev package |
| `make: command not found` | Missing build tools | Install make and build-essential |
| `connection refused` | Network/firewall | Check network connectivity and firewall |

---

**Still having issues?** 🤔 

Create a GitHub issue with detailed information and we'll help you get it working!