# File: scripts/test.sh
#!/bin/bash

set -e

echo "Starting test services..."
docker-compose -f docker-compose.test.yml up -d

echo "Waiting for services to be ready..."
sleep 15  # Increased sleep to ensure services are ready

echo "Running tests..."
export DB_HOST=localhost
export DB_PORT=5432  # Matches docker-compose.test.yml
export DB_USER=postgres  # Matches docker-compose.test.yml
export DB_PASSWORD=postgres  # Matches docker-compose.test.yml
export DB_NAME=test_db  # Matches docker-compose.test.yml
export DB_SSL_MODE=disable
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=""
export REDIS_DB=0
export JWT_SECRET=test-secret-key
export LOG_LEVEL=error

go test -v -tags=integration ./...

echo "Generating coverage report..."
go tool cover -html=coverage.out -o coverage.html

echo "Stopping test services..."
docker-compose -f docker-compose.test.yml down

echo "Tests completed successfully!"
