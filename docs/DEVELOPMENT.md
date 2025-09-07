# Development Guide 🛠️

Complete guide for developers who want to contribute to Disruptor or understand the codebase architecture.

## Table of Contents

- [Quick Development Setup](#quick-development-setup)
- [Project Architecture](#project-architecture)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Database Development](#database-development)
- [Adding Features](#adding-features)
- [Code Style](#code-style)
- [Debugging](#debugging)
- [Contributing](#contributing)

## Quick Development Setup ⚡

### Prerequisites

```bash
# macOS
brew install go opus pkg-config

# Ubuntu/Debian
sudo apt install golang-go libopus-dev pkg-config

# Arch Linux
sudo pacman -S go opus pkgconf
```

### Setup

```bash
# Clone repository
git clone https://github.com/XanderD99/disruptor.git
cd disruptor

# Install dependencies
go mod download

# Build binaries
make build build-migrate

# Set up development environment
export CONFIG_TOKEN="your_dev_bot_token"
export CONFIG_DATABASE_DSN="file::memory:?cache=shared"  # In-memory for dev
export CONFIG_LOGGING_LEVEL="debug"

# Run in development mode
./output/bin/disruptor
```

## Project Architecture 🏗️

```ansii
disruptor/
├── cmd/                    # Application entrypoints
│   ├── disruptor/         # Main bot application
│   │   ├── main.go       # Application entry point
│   │   └── config.go     # Configuration structure
│   └── migrate/           # Database migration tool
│       ├── main.go       # Migration runner
│       └── migrations/   # SQL migration files
├── internal/              # Private application logic
│   ├── commands/         # Discord slash commands
│   │   ├── play.go      # /play command
│   │   ├── interval.go  # /interval command
│   │   ├── chance.go    # /chance command
│   │   ├── weight.go    # /weight command
│   │   ├── next.go      # /next command
│   │   └── disconnect.go# /disconnect command
│   ├── disruptor/        # Core bot logic
│   │   ├── config.go    # Bot configuration
│   │   ├── session.go   # Discord session wrapper
│   │   └── interfaces.go# Core interfaces
│   ├── handlers/         # Event handlers
│   │   └── handlers/    # Scheduler handlers
│   ├── listeners/        # Discord event listeners
│   │   ├── guild_ready.go
│   │   ├── guild_join.go
│   │   └── guild_leave.go
│   ├── models/           # Database models (Bun ORM)
│   │   ├── guild.go     # Guild model
│   │   └── channel.go   # Channel model
│   ├── scheduler/        # Task scheduling system
│   │   ├── scheduler.go # Main scheduler
│   │   ├── manager.go   # Schedule manager
│   │   └── handlers/    # Scheduled task handlers
│   ├── metrics/          # Prometheus metrics
│   ├── middlewares/      # HTTP/Discord middlewares
│   ├── util/            # Shared utilities
│   └── http/            # HTTP server (metrics, health)
├── pkg/                  # Public reusable packages
│   ├── logging/         # Structured logging
│   └── processes/       # Process management
└── docs/                # Documentation
```

### Key Components

#### 1. **Commands** (`internal/commands/`)

Each Discord slash command is a separate module implementing the `disruptor.Command` interface:

```go
type Command interface {
    Load(handler.Router)           // Register command handler
    Options() discord.SlashCommandCreate  // Define command structure
}
```

#### 2. **Scheduler** (`internal/scheduler/`)

Manages timed disruptions using worker pools and jitter:

- **Manager**: Coordinates multiple schedules
- **Scheduler**: Handles individual schedule execution
- **Handlers**: Execute scheduled tasks (e.g., random voice joins)

#### 3. **Models** (`internal/models/`)

Bun ORM models for database entities:

- **Guild**: Server-specific settings (interval, chance)
- **Channel**: Channel-specific settings (weight)

#### 4. **Session** (`internal/disruptor/session.go`)

Wraps Discord session with additional functionality:

- Caching layer
- Logger integration
- Metrics collection

## Development Workflow 🔄

### 1. **Local Development**

```bash
# Start with in-memory database
export CONFIG_DATABASE_DSN="file::memory:?cache=shared"
export CONFIG_LOGGING_LEVEL="debug"

# Run the bot
make build && ./output/bin/disruptor

# Or run with hot reload (if using air)
air
```

### 2. **Working with Database**

```bash
# Use persistent database for development
export CONFIG_DATABASE_DSN="file:./dev.db?cache=shared"

# Run migrations manually
./output/bin/migrate up

# Reset database
rm dev.db && ./output/bin/migrate up
```

### 3. **Adding New Commands**

```bash
# 1. Create new command file
touch internal/commands/mycommand.go

# 2. Implement Command interface
# 3. Register in main.go
# 4. Test with Discord slash command
```

## Testing 🧪

### Current Test Status

- **Unit tests**: Infrastructure set up, but no tests implemented yet
- **Integration tests**: Not implemented
- **Docker build tests**: Included in CI/CD pipeline

### Running Tests

```bash
# Run all tests (currently none)
make test

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/commands/...
```

### Writing Tests

Create test files alongside source files:

```go
// internal/commands/play_test.go
package commands

import (
    "testing"
)

func TestPlayCommand(t *testing.T) {
    // Test implementation
}
```

## Database Development 🗄️

### Migration System

```bash
# Create new migration
mkdir -p cmd/migrate/migrations
# Create files: YYYYMMDDHHMMSS_description.go

# Run migrations
./output/bin/migrate up

# Migration template
package migrations

import (
    "context"
    "github.com/uptrace/bun"
)

func init() {
    Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
        // Migration up logic
        return nil
    }, func(ctx context.Context, db *bun.DB) error {
        // Migration down logic
        return nil
    })
}
```

### Working with Models

```go
// Add new model field
type Guild struct {
    ID       snowflake.ID `bun:"id,pk"`
    NewField string       `bun:"new_field"`
}

// Create migration for schema change
// Update model validation
// Test with development database
```

## Adding Features 🚀

### 1. **New Slash Command**

```go
// 1. Create internal/commands/feature.go
package commands

type feature struct {
    db *bun.DB
}

func Feature(db *bun.DB) disruptor.Command {
    return feature{db: db}
}

func (f feature) Load(r handler.Router) {
    r.SlashCommand("/feature", f.handle)
}

func (f feature) Options() discord.SlashCommandCreate {
    return discord.SlashCommandCreate{
        Name: "feature",
        Description: "New feature description",
        // Add options...
    }
}

func (f feature) handle(d discord.SlashCommandInteractionData, event *handler.CommandEvent) error {
    // Implementation
}

// 2. Register in cmd/disruptor/main.go
session.LoadCommands(
    commands.Feature(db),
    // ... other commands
)
```

### 2. **New Scheduler Handler**

```go
// 1. Create internal/scheduler/handlers/new_handler.go
func NewFeatureHandler(session *disruptor.Session, db *bun.DB) scheduler.HandleFunc {
    return func(ctx context.Context) error {
        // Handler implementation
    }
}

// 2. Register in scheduler initialization
```

### 3. **New Metrics**

```go
// Add to internal/metrics/
var featureMetric = promauto.NewCounterVec(
    prometheus.CounterOpts{
        Name: "disruptor_feature_total",
        Help: "Total feature executions",
    },
    []string{"guild_id", "status"},
)
```

## Code Style 📝

### Go Guidelines

- **Formatting**: Use `go fmt` (automatically applied)
- **Linting**: Use `go vet` (avoid `make lint` due to config issues)
- **Imports**: Group standard, external, and internal imports
- **Error handling**: Always handle errors explicitly
- **Naming**: Use Go naming conventions (camelCase, PascalCase)

### Project Conventions

```go
// Good: Descriptive function names
func determineVoiceChannelID(ctx context.Context, session *disruptor.Session, guild models.Guild) (snowflake.ID, error)

// Good: Error wrapping
return fmt.Errorf("failed to get channels for guild %s: %w", guild.ID, err)

// Good: Context usage
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()
```

### Database Conventions

```go
// Model naming: Singular, PascalCase
type Guild struct {
    ID snowflake.ID `bun:"id,pk" validate:"required"`
}

// Default constructors
func DefaultGuild(id snowflake.ID) *Guild {
    return &Guild{ID: id, Interval: time.Hour, Chance: 50.0}
}
```

## Debugging 🐛

### Logging

```go
// Use structured logging
session.LoggerInfoContext(ctx, "Processing guild",
    slog.Any("guild.id", guild.ID),
    slog.Int("channels", len(channels)),
)

// Debug level for detailed info
session.LoggerDebugContext(ctx, "Channel selection",
    slog.Any("available", available),
    slog.Any("weights", weights),
)
```

### Common Debug Patterns

```bash
# Enable debug logging
export CONFIG_LOGGING_LEVEL="debug"

# Check database state
sqlite3 dev.db ".tables"
sqlite3 dev.db "SELECT * FROM guilds;"

# Monitor metrics
curl http://localhost:8080/metrics

# Check Discord permissions
# Bot needs: Connect, Speak, Use Slash Commands, Manage Guild
```

### Debug Discord Issues

```bash
# Check bot permissions in guild
# Verify soundboard sounds exist
# Ensure voice channels have users
# Check slash command registration (up to 1 hour delay)
```

## Contributing 🤝

### Before You Start

1. **Read**: [CONTRIBUTING.md](../CONTRIBUTING.md)
2. **Discuss**: Open an issue for new features
3. **Fork**: Create your feature branch
4. **Develop**: Follow this guide
5. **Test**: Ensure code works
6. **Document**: Update relevant docs
7. **Submit**: Create pull request

### Pull Request Checklist

- [ ] Code follows project conventions
- [ ] Tests written (when applicable)
- [ ] Documentation updated
- [ ] `go vet ./...` passes
- [ ] `go fmt` applied
- [ ] Commit messages are descriptive
- [ ] No sensitive information in commits

### Development Tips

- **Start small**: Begin with minor improvements
- **Ask questions**: Use GitHub discussions
- **Test thoroughly**: Try different Discord server configurations
- **Document changes**: Update relevant documentation
- **Follow patterns**: Match existing code style and structure

---

**Happy coding!** 🎉 The Disruptor community appreciates your contributions to the delightful chaos!
