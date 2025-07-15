#!/bin/bash
set -e

echo "🔍 Running linters..."

# Go formatting
echo "Checking Go formatting..."
if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
    echo "❌ Code is not formatted. Run 'gofmt -s -w .'"
    exit 1
fi

# Go vet
echo "Running go vet..."
go vet ./...

# golangci-lint (if installed)
if command -v golangci-lint &> /dev/null; then
    echo "Running golangci-lint..."
    golangci-lint run --config=ci/config/.golangci.yml
else
    echo "⚠️ golangci-lint not found, skipping..."
fi

echo "✅ All linting checks passed!"
