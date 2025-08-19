# Disruptor Discord Bot Development Instructions

**ALWAYS follow these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.**

Disruptor is a Go-based Discord bot that randomly joins voice channels to play soundboard sounds. It uses Discord's native soundboard API, Lavalink for audio streaming, SQLite with Bun ORM for data storage, and comprehensive CI/CD with GitHub Actions.

## Working Effectively

### Prerequisites and System Setup
- Install system dependencies:
  ```bash
  sudo apt-get update
  sudo apt-get install -y libopus-dev pkg-config
  ```
- Ensure Go 1.24+ is installed (`go version` should show 1.24.x or later)

### Bootstrap, Build, and Test
- Download dependencies:
  ```bash
  go mod download
  ```
  Expected time: <1 second (cached), ~12 seconds (first time). Set timeout to 30+ seconds.

- Build the application:
  ```bash
  make build
  ```
  Expected time: ~1 second (cached), ~34 seconds (clean build). **NEVER CANCEL.** Set timeout to 90+ seconds.

- Run tests (note: currently no tests exist, but command works):
  ```bash
  make test
  ```
  Expected time: <1 second. Set timeout to 30+ seconds.

- Run linting and formatting:
  ```bash
  go vet ./...        # Takes <1 second
  go fmt ./...        # Takes <1 second
  ```
  **DO NOT use `make lint`** - the golangci-lint config has path issues (.yaml vs .yml) and version compatibility problems.

### Running the Application
- **NEVER run the bot without proper Discord configuration** - it requires:
  ```bash
  export CONFIG_TOKEN=your_discord_bot_token
  # Optional persistent database:
  export CONFIG_DATABASE_DSN=file:./disruptor.db?cache=shared
  ./output/bin/disruptor
  ```
- Application will fail without `CONFIG_TOKEN` environment variable
- Default database is in-memory SQLite (`file::memory:?cache=shared`)
- **Cannot test actual Discord functionality** without a valid bot token and server setup

### Docker Build and Deployment
- Build Docker image:
  ```bash
  make docker-build
  ```
  OR manually:
  ```bash
  docker build --target final -t disruptor:latest -f ./ci/Dockerfile .
  ```
  Expected time: 3-5 minutes depending on network. **NEVER CANCEL.** Set timeout to 10+ minutes.

## Validation

### Manual Validation After Changes
Since this is a Discord bot, **ALWAYS validate changes using these approaches**:

1. **Build Validation**: Always run `make build` after code changes to ensure compilation
2. **Static Analysis**: Run `go vet ./...` to catch potential issues
3. **Configuration Testing**: 
   - Test without token: `./output/bin/disruptor` should show "required environment variable CONFIG_TOKEN is not set"
   - Test with token: `export CONFIG_TOKEN="test" && timeout 5s ./output/bin/disruptor` should start logging process manager initialization
   - Test database configurations: Both `CONFIG_DATABASE_DSN=file::memory:?cache=shared` and `CONFIG_DATABASE_DSN=file:./test.db?cache=shared` should work
4. **Database Schema**: If modifying models, ensure the app starts with both in-memory and file-based SQLite DSNs
5. **Import Validation**: Check that all imports are properly organized with `go fmt ./...`

### Testing Limitations
- **No unit tests exist currently** - the repository has test infrastructure but no actual test files
- **Cannot test Discord functionality** without bot token and Discord server setup
- **Cannot test Lavalink integration** without Lavalink server running
- **Focus on compilation, static analysis, and configuration validation**

### Required Pre-commit Validation
Before committing changes, **ALWAYS run**:
```bash
go fmt ./...    # Format code
go vet ./...    # Static analysis
make build      # Ensure compilation
```
Do NOT run `make lint` due to configuration issues.

## Common Tasks and File Locations

### Key Directories and Files
- `cmd/disruptor/main.go` - Application entry point
- `internal/commands/` - Discord slash commands
- `internal/handlers/` - Discord event handlers  
- `internal/config/` - Configuration management
- `internal/models/` - Database models (Bun ORM)
- `internal/scheduler/` - Audio scheduling logic
- `ci/` - CI/CD configuration and scripts
- `makefile` - Build automation

### Environment Configuration
All configuration via environment variables (see `internal/config/README.md`):
- `CONFIG_TOKEN` (required) - Discord bot token
- `CONFIG_DATABASE_DSN` - Database connection string
- `CONFIG_LOGGING_LEVEL` - Log verbosity (debug, info, warn, error)
- `CONFIG_LAVALINK_*` - Lavalink audio server settings
- Full list in `configs/.env.example`

### Build System Notes
- **Makefile targets**: `build`, `test`, `run`, `docker-build`, `help`
- **Output directory**: `output/bin/` for compiled binaries
- **CGO enabled**: Required for SQLite driver
- **Build flags**: Automatic version injection via git describe

### CI/CD Pipeline
GitHub Actions runs on push/PR to main/develop:
1. **Test job**: Downloads deps, runs tests with coverage (~12s + ~6s)
2. **Lint job**: Runs golangci-lint with custom config (~6s)  
3. **Build job**: Compiles application and uploads artifacts (~34s)
4. **Docker job**: Builds and tests Docker image (~3-5 minutes)

### Known Issues and Workarounds
1. **golangci-lint not installed**: `make lint` will fail with "command not found" - use `go vet` instead
2. **golangci-lint config**: When installed, has path mismatch (.yaml vs .yml) - use `go vet` instead
3. **golangci-lint version**: When installed, may be incompatible with Go 1.24 - use `go vet` instead  
4. **Discord token requirement**: App cannot run without valid Discord bot token
5. **No unit tests**: Test infrastructure exists but no actual tests written

### Performance Expectations
- **Module download**: <1 second (cached), ~12 seconds (first time)
- **Build time**: ~1 second (cached), ~34 seconds (clean build)
- **Test execution**: <1 second (no actual tests)
- **Static analysis**: <1 second
- **Docker build**: 3-5 minutes
- **All operations under 1 minute except Docker builds**

**CRITICAL TIMING NOTES:**
- **NEVER CANCEL** any build or test operation under 2 minutes
- Set minimum 90-second timeout for `make build`
- Set minimum 30-second timeout for `go mod download` and tests
- Set minimum 10-minute timeout for Docker builds