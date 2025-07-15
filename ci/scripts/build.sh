#!/bin/bash
set -e

echo "🏗️ Building Discord Bot..."

# Run tests first
echo "🧪 Running tests..."
go test -v ./... -coverprofile=coverage.out

# Build the application
echo "📦 Building application..."
make build

echo "✅ Build completed successfully!"
