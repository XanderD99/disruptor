FROM golang:1.25-alpine AS init

WORKDIR /app
COPY go.mod go.sum ./

# Install dependencies and build tools
RUN apk add --no-cache gcc musl-dev make git bash pkgconfig opus-dev bind-tools

# Install Go dependencies
RUN go mod download

COPY . .

ENV CGO_ENABLED=1

# Test step - Run unit tests and fail build if they fail
FROM init AS tester
RUN go test -v ./... -coverprofile=coverage.out || (echo "❌ Unit tests failed! Build stopped." && exit 1)

# Build step - Only runs if tests pass
FROM init AS build

RUN make build
RUN make build-migrate

FROM alpine:3.19 AS final

RUN apk add --no-cache opus ffmpeg

# Create app directory and set proper permissions
RUN mkdir -p /app && chmod 755 /app
WORKDIR /app

# Copy only the built binary from the builder stage
COPY --from=build /app/output/bin/disruptor /bin/disruptor
COPY --from=build /app/output/bin/migrate /bin/migrate
# Copy migration files needed by the migrate binary
COPY --from=build /app/cmd/migrate/migrations/ /app/cmd/migrate/migrations/

# Create a non-root user to run the application
RUN adduser -D -H -h /app disruptor && chown -R disruptor:disruptor /app
USER disruptor

ENTRYPOINT ["/bin/disruptor"]
