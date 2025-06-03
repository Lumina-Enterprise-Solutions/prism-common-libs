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
	Consul      ConsulConfig // Tambahkan ini
}

type JWTConfig struct {
	Secret         string // Akan diisi dari Vault atau env
	ExpirationTime int
	// VaultPath      string // Opsional: path ke secret di Vault
	// VaultKey       string // Opsional: key dari secret di Vault
}

type ConsulConfig struct {
	Address string
	Token   string
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
	// Muat JWT Secret dari Vault atau fallback ke environment variable
	// jwtSecret := getEnvString("JWT_SECRET", "your-secret-key-default-fallback") // Default fallback

	// Uncomment dan sesuaikan jika ingin load dari Vault:
	/*
	   vaultClient, err := secrets.NewVaultClient() // Asumsikan VAULT_ADDR dan VAULT_TOKEN ada di env
	   if err != nil {
	       // Mungkin log warning dan lanjutkan dengan env var, atau return error
	       fmt.Printf("Warning: Failed to create Vault client: %v. Falling back to JWT_SECRET env var.\n", err)
	   } else {
	       // Path dan key ini harus sesuai dengan yang Anda simpan di Vault
	       vaultPath := getEnvString("JWT_VAULT_PATH", "prism/auth-service/jwt")
	       vaultKey := getEnvString("JWT_VAULT_KEY", "secret")

	       retrievedSecret, err := vaultClient.ReadSecret(vaultPath, vaultKey)
	       if err != nil {
	           fmt.Printf("Warning: Failed to read JWT secret from Vault (%s, key %s): %v. Falling back to JWT_SECRET env var.\n", vaultPath, vaultKey, err)
	       } else {
	           jwtSecret = retrievedSecret
	           fmt.Println("Successfully loaded JWT secret from Vault.")
	       }
	   }
	*/
	config := &Config{
		Environment: getEnvString("ENVIRONMENT", "development"),
		TenantID:    getEnvString("TENANT_ID", "default"),
		Consul: ConsulConfig{
			Address: getEnvString("CONSUL_ADDRESS", "http://localhost:8500"),
			Token:   getEnvString("CONSUL_TOKEN", ""),
		},
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
			Secret:         getEnvString("JWT_SECRET", "your-secret-key"), // Ini akan diganti jika Vault berhasil
			ExpirationTime: getEnvInt("JWT_EXPIRATION", DefaultJWTExpiration),
			// VaultPath:      getEnvString("JWT_VAULT_PATH", "prism/auth-service/jwt"),
			// VaultKey:       getEnvString("JWT_VAULT_KEY", "secret"),
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
