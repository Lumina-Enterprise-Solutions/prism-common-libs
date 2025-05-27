# Prism Common Libraries

[![Go Report Card](https://goreportcard.com/badge/github.com/Lumina-Enterprise-Solutions/prism-common-libs)](https://goreportcard.com/report/github.com/Lumina-Enterprise-Solutions/prism-common-libs)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI/CD Pipeline](https://github.com/Lumina-Enterprise-Solutions/prism-common-libs/actions/workflows/ci.yml/badge.svg)](https://github.com/Lumina-Enterprise-Solutions/prism-common-libs/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8.svg)](https://golang.org/dl/)

`prism-common-libs` is a Go library providing common utilities and components for building scalable, multi-tenant applications with the Prism ERP system. It includes middleware, database connectivity, caching, logging, and more.

## Features

- **Middleware**: Authentication, CORS, request ID, and tenant handling for Gin-based applications.
- **Database**: PostgreSQL integration with tenant isolation using schema-based separation.
- **Caching**: Redis client for efficient data caching and retrieval.
- **Utilities**: Validation, response formatting, and cryptographic functions.
- **Logging**: Configurable logging with Logrus.
- **Models**: Base models for users, roles, and tenants with GORM integration.
- **Testing**: Comprehensive unit and integration tests with Docker Compose for test environments.
- **Benchmarking**: Performance benchmarks for database and Redis operations.

## Project Structure

```
prism-common-libs/
├── scripts/                # Utility scripts for linting and testing
├── pkg/
│   ├── middleware/         # Gin middleware for auth, CORS, etc.
│   ├── database/           # PostgreSQL database connectivity
│   ├── utils/              # Utility functions (crypto, validation, response)
│   ├── integration/        # Integration tests
│   ├── benchmark/          # Performance benchmarks
│   ├── logger/             # Logging utilities
│   ├── models/             # GORM models
│   ├── cache/              # Redis caching
│   └── config/             # Configuration management
├── .mdignore               # Files to ignore in markdown generation
├── docker-compose.test.yml # Docker Compose for test environment
└── Makefile                # Build and test automation
```

## Prerequisites

- Go 1.20 or higher
- Docker and Docker Compose (for integration tests)
- PostgreSQL (for database operations)
- Redis (for caching)

## Installation

```bash
go get github.com/Lumina-Enterprise-Solutions/prism-common-libs
```

## Usage

### Configuration

The library uses environment variables for configuration. Example configuration:

```go
package main

import (
	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	// Use cfg for database, redis, etc.
}
```

### Middleware Example

```go
package main

import (
	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())
	r.Use(middleware.TenantMiddleware())
	r.Use(middleware.RequireAuth(cfg.JWT))

	r.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Authenticated"})
	})

	r.Run()
}
```

### Database Example

```go
package main

import (
	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/database"
	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/models"
)

func main() {
	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		panic(err)
	}

	// Run migrations
	db.DB.AutoMigrate(&models.User{}, &models.Role{}, &models.Tenant{})

	// Create a user
	user := models.User{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Status:    "active",
	}
	db.DB.Create(&user)
}
```

## Development

### Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/Lumina-Enterprise-Solutions/prism-common-libs.git
   cd prism-common-libs
   ```

2. Install dependencies:
   ```bash
   make deps
   ```

3. Install development tools:
   ```bash
   make install-tools
   ```

### Running Tests

Run unit and integration tests:
```bash
make test
```

Run only unit tests:
```bash
make test-unit
```

Run only integration tests (requires Docker):
```bash
make test-integration
```

### Linting and Formatting

Run linters:
```bash
make lint
```

Format code:
```bash
make format
```

### Generating Coverage Report

```bash
make coverage-html
```

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Commit your changes (`git commit -m "Add your feature"`)
4. Push to the branch (`git push origin feature/your-feature`)
5. Create a Pull Request

Please ensure your code passes linting and tests before submitting.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
