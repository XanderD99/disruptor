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
FROM init AS builder

RUN go build -o /app/output/bin/disruptor --ldflags="-s -w" -trimpath ./cmd/disruptor
RUN go build -o /app/output/bin/migrate --ldflags="-s -w" -trimpath ./cmd/migrate

FROM alpine:3.19 AS final

RUN apk add --no-cache opus ffmpeg

# Copy only the built binary from the builder stage
COPY --from=builder /app/output/bin/disruptor /bin/disruptor
COPY --from=builder /app/output/bin/migrate /bin/migrate

# Create a non-root user to run the application
RUN adduser -D -H -h /app disruptor
USER disruptor

ENTRYPOINT ["/bin/disruptor"]
