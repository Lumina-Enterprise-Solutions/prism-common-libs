.PHONY: test test-integration test-unit lint format build clean deps help

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download dependencies
	go mod download
	go mod verify

test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	go test -v -race -coverprofile=coverage.out ./...

test-integration: ## Run integration tests with services
	docker-compose -f docker-compose.test.yml up -d
	sleep 10
	go test -v -tags=integration ./...
	docker-compose -f docker-compose.test.yml down

lint: ## Run linters
	golangci-lint run

format: ## Format code
	gofmt -s -w .
	goimports -w .

build: ## Build the project
	go build -v ./...

clean: ## Clean build artifacts
	go clean -cache
	rm -f coverage.out coverage.html

coverage-html: test-unit ## Generate HTML coverage report
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
