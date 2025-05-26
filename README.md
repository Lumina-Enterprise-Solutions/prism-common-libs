# Prism Common Libraries

[![CI/CD Pipeline](https://github.com/Lumina-Enterprise-Solutions/prism-common-libs/actions/workflows/ci.yml/badge.svg)](https://github.com/Lumina-Enterprise-Solutions/prism-common-libs/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.24.3-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/Lumina-Enterprise-Solutions/prism-common-libs)](https://goreportcard.com/report/github.com/Lumina-Enterprise-Solutions/prism-common-libs)

A comprehensive set of shared libraries and utilities for the Prism ERP ecosystem, providing common functionality across microservices including authentication, database operations, caching, logging, and middleware components.

## 🚀 Features

- **Multi-tenant Architecture**: Built-in tenant isolation and management
- **Authentication & Authorization**: JWT-based authentication with role-based access control
- **Database Integration**: PostgreSQL with GORM ORM and tenant-aware operations
- **Redis Caching**: High-performance caching layer with JSON serialization
- **Structured Logging**: JSON-formatted logging with configurable levels
- **HTTP Middleware**: CORS, authentication, tenant isolation, and request tracking
- **Validation Utilities**: Input validation with detailed error formatting
- **Security Utilities**: Cryptographic functions and secure random generation
- **Configuration Management**: Environment-based configuration with sensible defaults

## 📦 Installation

```bash
go get github.com/Lumina-Enterprise-Solutions/prism-common-libs
```

## 🏗️ Project Structure

```
prism-common-libs/
├── pkg/
│   ├── cache/          # Redis caching operations
│   ├── config/         # Configuration management
│   ├── database/       # PostgreSQL database connections
│   ├── logger/         # Structured logging utilities
│   ├── middleware/     # HTTP middleware components
│   ├── models/         # Common data models
│   └── utils/          # Utility functions
├── .github/
│   └── workflows/      # CI/CD pipeline configuration
└── docs/              # Documentation
```

## 🔧 Quick Start

### 1. Configuration

Create a `.env` file or set environment variables:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=prism_erp
DB_USER=prism
DB_PASSWORD=your_password
DB_SSL_MODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
REDIS_DB=0

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRATION=3600

# Server Configuration
SERVER_PORT=8080
SERVER_READ_TIMEOUT=10
SERVER_WRITE_TIMEOUT=10

# Application Configuration
ENVIRONMENT=development
TENANT_ID=default
LOG_LEVEL=info
```

### 2. Basic Usage

```go
package main

import (
    "log"
    "github.com/gin-gonic/gin"
    "github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/config"
    "github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/database"
    "github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/cache"
    "github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/middleware"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }

    // Initialize database
    db, err := database.NewPostgresConnection(cfg.Database)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // Initialize Redis cache
    redisClient := cache.NewRedisClient(cfg.Redis)

    // Setup Gin router with middleware
    r := gin.Default()
    r.Use(middleware.CORS())
    r.Use(middleware.RequestID())
    r.Use(middleware.TenantMiddleware())

    // Protected routes
    protected := r.Group("/api")
    protected.Use(middleware.RequireAuth(cfg.JWT))

    // Your routes here
    protected.GET("/users", getUsersHandler)

    r.Run(fmt.Sprintf(":%d", cfg.Server.Port))
}
```

## 📚 Package Documentation

### Config Package

Manages application configuration with environment variable support and sensible defaults.

```go
import "github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/config"

cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

// Access configuration
dbDSN := cfg.Database.DSN()
redisAddr := cfg.Redis.Address()
```

### Database Package

PostgreSQL database connection with GORM integration and multi-tenant support.

```go
import "github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/database"

db, err := database.NewPostgresConnection(cfg.Database)
if err != nil {
    log.Fatal(err)
}

// Use with tenant isolation
tenantDB := db.WithTenant("tenant_123")
```

### Cache Package

Redis-based caching with JSON serialization support.

```go
import "github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/cache"

client := cache.NewRedisClient(cfg.Redis)

// Set cache with expiration
err := client.Set(ctx, "user:123", userData, time.Hour)

// Get from cache
var user User
err := client.Get(ctx, "user:123", &user)

// Check existence
exists, err := client.Exists(ctx, "user:123")

// Delete from cache
err := client.Delete(ctx, "user:123")
```

### Logger Package

Structured JSON logging with configurable levels.

```go
import "github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/logger"

// Simple logging
logger.Info("Application started")
logger.Error("Database connection failed")

// Structured logging with fields
logger.WithFields(logrus.Fields{
    "user_id": "123",
    "action":  "login",
}).Info("User logged in")
```

### Middleware Package

HTTP middleware for Gin framework including CORS, authentication, and tenant management.

```go
import "github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/middleware"

r := gin.Default()

// Apply middleware
r.Use(middleware.CORS())
r.Use(middleware.RequestID())
r.Use(middleware.TenantMiddleware())

// Protected routes
protected := r.Group("/api")
protected.Use(middleware.RequireAuth(cfg.JWT))
```

### Models Package

Common data models with UUID primary keys and soft deletes.

```go
import "github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/models"

type CustomModel struct {
    models.BaseModel
    Name string `json:"name"`
    // Additional fields
}
```

### Utils Package

#### Response Utilities

```go
import "github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/utils"

// Success response
utils.SuccessResponse(c, "User created successfully", user)

// Error response
utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input", err)

// Validation error response
utils.ValidationErrorResponse(c, validationErrors)
```

#### Validation Utilities

```go
// Format validation errors
errors := utils.FormatValidationErrors(validationErr)
```

#### Crypto Utilities

```go
// Generate random bytes
randomBytes := utils.GenerateRandomBytes(32)

// Generate random hex string
randomString := utils.GenerateRandomString(16)
```

## 🔐 Authentication & Authorization

The library provides JWT-based authentication with the following features:

- **Token Generation**: Create signed JWT tokens with user claims
- **Token Validation**: Middleware for validating JWT tokens
- **Multi-tenant Support**: Tenant-aware authentication
- **Role-based Access Control**: Flexible permission system

### JWT Token Structure

```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "tenant_id": "tenant_123",
  "roles": ["admin", "user"],
  "exp": 1640995200,
  "iat": 1640991600
}
```

## 🏢 Multi-tenant Architecture

The library supports multi-tenant applications with:

- **Tenant Isolation**: Database schema separation per tenant
- **Tenant Context**: Automatic tenant detection from headers
- **Tenant-aware Models**: Built-in tenant filtering
- **Tenant Configuration**: Per-tenant configuration support

## 🐳 Docker Support

```dockerfile
FROM golang:1.24.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

## 🧪 Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html
```

## 🚀 CI/CD Pipeline

The project includes a comprehensive CI/CD pipeline with:

- **Automated Testing**: Unit tests with coverage reporting
- **Security Scanning**: Gosec and Trivy vulnerability scanning
- **Code Quality**: golangci-lint integration
- **Docker Build**: Multi-stage Docker builds
- **Container Registry**: GitHub Container Registry integration

## 📋 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ENVIRONMENT` | Application environment | `development` |
| `TENANT_ID` | Default tenant ID | `default` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_NAME` | Database name | `prism_erp` |
| `DB_USER` | Database username | `prism` |
| `DB_PASSWORD` | Database password | `prism123` |
| `DB_SSL_MODE` | Database SSL mode | `disable` |
| `REDIS_HOST` | Redis host | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `REDIS_PASSWORD` | Redis password | `redis123` |
| `REDIS_DB` | Redis database number | `0` |
| `JWT_SECRET` | JWT signing secret | `your-secret-key` |
| `JWT_EXPIRATION` | JWT expiration time (seconds) | `3600` |
| `SERVER_PORT` | Server port | `8080` |
| `SERVER_READ_TIMEOUT` | Server read timeout (seconds) | `10` |
| `SERVER_WRITE_TIMEOUT` | Server write timeout (seconds) | `10` |
| `LOG_LEVEL` | Logging level | `info` |

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/Lumina-Enterprise-Solutions/prism-common-libs.git
   cd prism-common-libs
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Run tests**
   ```bash
   go test ./...
   ```

4. **Run linting**
   ```bash
   golangci-lint run
   ```

### Code Standards

- Follow Go best practices and idioms
- Write comprehensive tests for new features
- Document public APIs with clear examples
- Use meaningful commit messages
- Ensure all CI checks pass

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: [Wiki](https://github.com/Lumina-Enterprise-Solutions/prism-common-libs/wiki)
- **Issues**: [GitHub Issues](https://github.com/Lumina-Enterprise-Solutions/prism-common-libs/issues)
- **Discussions**: [GitHub Discussions](https://github.com/Lumina-Enterprise-Solutions/prism-common-libs/discussions)

## 🚀 Roadmap

- [ ] **GraphQL Support**: Add GraphQL middleware and utilities
- [ ] **Message Queuing**: RabbitMQ/Kafka integration
- [ ] **Metrics & Monitoring**: Prometheus metrics collection
- [ ] **API Rate Limiting**: Redis-based rate limiting middleware
- [ ] **Event Sourcing**: Event store utilities
- [ ] **OpenAPI Integration**: Automatic API documentation generation
- [ ] **Health Checks**: Comprehensive health check endpoints
- [ ] **Circuit Breaker**: Resilience patterns implementation

## 🏆 Acknowledgments

- Built with ❤️ by the Lumina Enterprise Solutions team
- Powered by [Go](https://golang.org/), [Gin](https://gin-gonic.com/), [GORM](https://gorm.io/), and [Redis](https://redis.io/)
- Inspired by modern microservices architecture patterns

---

**Made with ☕ and 🧠 by [Lumina Enterprise Solutions](https://github.com/Lumina-Enterprise-Solutions)**
