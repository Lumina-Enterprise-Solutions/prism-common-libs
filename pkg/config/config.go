package config

import (
	"fmt"
	"os"
	"strconv"
)

type DatabaseConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	Secret         string
	ExpirationTime int
}

type ServerConfig struct {
	Port         int
	ReadTimeout  int
	WriteTimeout int
}

type Config struct {
	Environment string
	TenantID    string
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	Server      ServerConfig
}

const (
	DefaultDBPort        = 5432
	DefaultRedisPort     = 6379
	DefaultJWTExpiration = 3600 // 1 hour in seconds
	DefaultServerPort    = 8080
	DefaultReadTimeout   = 10 // seconds
	DefaultWriteTimeout  = 10 // seconds
)

func Load() (*Config, error) {
	config := &Config{
		Environment: getEnvString("ENVIRONMENT", "development"),
		TenantID:    getEnvString("TENANT_ID", "default"),
		Database: DatabaseConfig{
			Host:     getEnvString("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", DefaultDBPort),
			Database: getEnvString("DB_NAME", "prism_erp"),
			Username: getEnvString("DB_USER", "prism"),
			Password: getEnvString("DB_PASSWORD", "prism123"),
			SSLMode:  getEnvString("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnvString("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", DefaultRedisPort),
			Password: getEnvString("REDIS_PASSWORD", "redis123"),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:         getEnvString("JWT_SECRET", "your-secret-key"),
			ExpirationTime: getEnvInt("JWT_EXPIRATION", DefaultJWTExpiration),
		},
		Server: ServerConfig{
			Port:         getEnvInt("SERVER_PORT", DefaultServerPort),
			ReadTimeout:  getEnvInt("SERVER_READ_TIMEOUT", DefaultReadTimeout),
			WriteTimeout: getEnvInt("SERVER_WRITE_TIMEOUT", DefaultWriteTimeout),
		},
	}

	return config, nil
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.Username, d.Password, d.Database, d.SSLMode)
}

func (r RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
