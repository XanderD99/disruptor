# Installation Guide 🛠️

This guide provides detailed installation instructions for Disruptor on different platforms and deployment scenarios.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Platform-Specific Installation](#platform-specific-installation)
  - [Linux (Ubuntu/Debian)](#linux-ubuntudebian)
  - [Linux (CentOS/RHEL/Fedora)](#linux-centosrhelfedora)
  - [macOS](#macos)
  - [Windows](#windows)
- [Installation Methods](#installation-methods)
  - [From Source](#from-source)
  - [Using Docker](#using-docker)
  - [Pre-built Binaries](#pre-built-binaries)
- [Verification](#verification)
- [Next Steps](#next-steps)

---

## Prerequisites

### System Requirements
- **OS**: Linux, macOS, or Windows
- **RAM**: 256MB minimum, 512MB recommended
- **Storage**: 100MB for application + database storage
- **Network**: Internet connection for Discord API

### Required Software
- **Go**: Version 1.24 or later
- **Git**: For cloning the repository
- **Make**: For build automation (Linux/macOS)
- **pkg-config**: For C library dependencies
- **libopus-dev**: Audio codec library

### Optional Software
- **Docker**: For containerized deployment
- **systemd**: For service management (Linux)

---

## Platform-Specific Installation

### Linux (Ubuntu/Debian)

#### 1. Install System Dependencies
```bash
# Update package list
sudo apt update

# Install build tools and dependencies
sudo apt install -y git make pkg-config libopus-dev build-essential

# Install Go (if not already installed)
sudo apt install -y golang-go

# Verify Go version (should be 1.18+)
go version
```

#### 2. Install Latest Go (if needed)
```bash
# Download and install Go 1.24+
wget https://go.dev/dl/go1.24.4.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.4.linux-amd64.tar.gz

# Add to PATH (add to ~/.bashrc for permanent)
export PATH=$PATH:/usr/local/go/bin
```

### Linux (CentOS/RHEL/Fedora)

#### 1. Install System Dependencies
```bash
# For CentOS/RHEL (with EPEL)
sudo yum install -y git make pkgconfig opus-devel gcc

# For Fedora
sudo dnf install -y git make pkgconfig opus-devel gcc golang

# Install Go (CentOS/RHEL)
sudo yum install -y golang
```

### macOS

#### 1. Install Homebrew (if not installed)
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

#### 2. Install Dependencies
```bash
# Install required packages
brew install go git opus pkg-config

# Verify installation
go version
which git
```

#### 3. Install Xcode Command Line Tools (if needed)
```bash
xcode-select --install
```

### Windows

#### 1. Install Go
1. Download Go from https://golang.org/dl/
2. Run the installer (.msi file)
3. Follow installation wizard
4. Open Command Prompt and verify: `go version`

#### 2. Install Git
1. Download Git from https://git-scm.com/download/win
2. Run installer with default settings
3. Verify: `git --version`

#### 3. Install Build Tools
```powershell
# Using Chocolatey (recommended)
choco install make

# Or install Visual Studio Build Tools
# Download from: https://visualstudio.microsoft.com/downloads/#build-tools-for-visual-studio-2022
```

#### 4. Install pkg-config and Opus
```powershell
# Using vcpkg (for Windows)
git clone https://github.com/Microsoft/vcpkg.git
cd vcpkg
.\bootstrap-vcpkg.bat
.\vcpkg.exe install opus:x64-windows pkg-config:x64-windows
```

---

## Installation Methods

### From Source

This is the recommended method for development and customization.

#### 1. Clone Repository
```bash
git clone https://github.com/XanderD99/disruptor.git
cd disruptor
```

#### 2. Download Dependencies
```bash
go mod download
```

#### 3. Build Application
```bash
# Using Make (Linux/macOS)
make build

# Or using Go directly
go build -o output/bin/disruptor ./cmd/disruptor
```

#### 4. Verify Build
```bash
./output/bin/disruptor --help
```

### Using Docker

#### 1. Pull Pre-built Image (when available)
```bash
docker pull ghcr.io/xanderd99/disruptor:latest
```

#### 2. Or Build Locally
```bash
git clone https://github.com/XanderD99/disruptor.git
cd disruptor
make docker-build
```

#### 3. Run Container
```bash
docker run -d --name disruptor \
  -e CONFIG_TOKEN=your_discord_bot_token \
  -e CONFIG_DATABASE_DSN=file:./disruptor.db?cache=shared \
  -v $(pwd)/data:/app/data \
  disruptor:latest
```

### Pre-built Binaries

Pre-built binaries are available from GitHub Releases.

#### 1. Download Binary
```bash
# Linux
wget https://github.com/XanderD99/disruptor/releases/latest/download/disruptor-linux-amd64.tar.gz
tar -xzf disruptor-linux-amd64.tar.gz

# macOS
wget https://github.com/XanderD99/disruptor/releases/latest/download/disruptor-darwin-amd64.tar.gz
tar -xzf disruptor-darwin-amd64.tar.gz

# Windows
# Download disruptor-windows-amd64.zip from releases page
```

#### 2. Make Executable (Linux/macOS)
```bash
chmod +x disruptor
```

#### 3. Optional: Install Globally
```bash
# Linux/macOS
sudo mv disruptor /usr/local/bin/

# Windows: Add to PATH or move to system folder
```

---

## Verification

### 1. Version Check
```bash
./output/bin/disruptor --version
# or if installed globally:
disruptor --version
```

### 2. Configuration Test
```bash
# Should show missing token error
./output/bin/disruptor

# Expected output:
# Error loading configuration: env: required environment variable "CONFIG_TOKEN" is not set
```

### 3. Test with Token
```bash
export CONFIG_TOKEN="your_test_token"
timeout 5s ./output/bin/disruptor

# Should attempt to connect to Discord
```

### 4. Build Information
```bash
go version -m ./output/bin/disruptor
```

---

## Post-Installation Setup

### 1. Create Data Directory
```bash
mkdir -p /opt/disruptor/data
chmod 755 /opt/disruptor/data
```

### 2. Create Configuration File
```bash
# Create environment file
cat > /opt/disruptor/.env << EOF
CONFIG_TOKEN=your_discord_bot_token
CONFIG_DATABASE_DSN=file:/opt/disruptor/data/disruptor.db?cache=shared
CONFIG_LOGGING_LEVEL=info
CONFIG_LOGGING_PRETTY=false
EOF

chmod 600 /opt/disruptor/.env
```

### 3. Create Service (Linux with systemd)
```bash
sudo tee /etc/systemd/system/disruptor.service > /dev/null << EOF
[Unit]
Description=Disruptor Discord Bot
After=network.target

[Service]
Type=simple
User=disruptor
Group=disruptor
WorkingDirectory=/opt/disruptor
EnvironmentFile=/opt/disruptor/.env
ExecStart=/opt/disruptor/disruptor
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Create user
sudo useradd -r -s /bin/false disruptor
sudo chown -R disruptor:disruptor /opt/disruptor

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable disruptor
sudo systemctl start disruptor
```

---

## Troubleshooting Installation

### Go Version Issues
```bash
# Check Go version
go version

# If too old, install latest:
# Linux: Use Go website installer
# macOS: brew upgrade go
# Windows: Download new installer
```

### Build Failures
```bash
# Clean and rebuild
make clean
go clean -cache
go mod download
make build
```

### Permission Issues (Linux)
```bash
# Fix ownership
sudo chown -R $USER:$USER /path/to/disruptor

# Fix executable permissions
chmod +x output/bin/disruptor
```

### Missing Dependencies
```bash
# Ubuntu/Debian
sudo apt install -y pkg-config libopus-dev

# macOS
brew install opus pkg-config

# Check pkg-config can find opus
pkg-config --modversion opus
```

### CGO Issues
```bash
# Ensure CGO is enabled
export CGO_ENABLED=1

# Check build environment
go env
```

---

## Performance Optimization

### 1. Binary Size Optimization
```bash
# Build with optimizations
go build -ldflags="-s -w" -o output/bin/disruptor ./cmd/disruptor
```

### 2. Runtime Optimization
```bash
# Set Go runtime variables
export GOMAXPROCS=2
export GOGC=100
```

### 3. Memory Limits (systemd)
```ini
# Add to service file
[Service]
MemoryMax=512M
MemoryHigh=256M
```

---

## Next Steps

After successful installation:

1. **Configure Discord Bot**: Follow [Discord Setup Guide](DISCORD_SETUP.md)
2. **Configure Application**: See [Configuration Guide](CONFIGURATION.md)
3. **Deploy**: Check [Deployment Guide](DEPLOYMENT.md)
4. **Quick Start**: Try [Quick Start Guide](QUICKSTART.md)

---

## Getting Help

- 📖 **Documentation**: Check other guides in `/docs/`
- 🐛 **Issues**: Report bugs on GitHub Issues
- 💬 **Discussions**: Use GitHub Discussions for questions
- 📧 **Security**: Email for security issues

---

**Installation complete!** 🎉 Your Disruptor bot is ready for configuration.