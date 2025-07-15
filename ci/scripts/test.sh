#!/bin/bash
set -e

echo "🧪 Running comprehensive tests..."

# Unit tests
echo "Running unit tests..."
go test -v ./... -coverprofile=coverage.out

# Integration tests (if you have them)
echo "Running integration tests..."
go test -v ./... -tags=integration

# Generate coverage report
echo "Generating coverage report..."
go tool cover -html=coverage.out -o coverage.html

echo "✅ All tests passed!"
