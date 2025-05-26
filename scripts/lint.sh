# File: scripts/lint.sh
#!/bin/bash

set -e

echo "Running golangci-lint..."
golangci-lint run --timeout=5m

echo "Running go fmt check..."
if [ -n "$(gofmt -l .)" ]; then
    echo "Code is not formatted. Please run 'make format'"
    exit 1
fi

echo "Running go vet..."
go vet ./...

echo "Running go mod verify..."
go mod verify

echo "Linting completed successfully!"
