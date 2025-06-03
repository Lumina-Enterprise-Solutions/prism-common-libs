package config

import (
	"fmt"
	"os"
	"strconv"

	// Jika Anda memisahkan consul client untuk KV
	// "github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/consulkv"
	consulapi "github.com/hashicorp/consul/api" // Gunakan API Consul langsung
)

// Variabel global untuk Consul client, diinisialisasi sekali
var consulKVClient *consulapi.Client

// initConsulKVClient inisialisasi Consul client untuk KV store
func initConsulKVClient(consulAddress string) error {
	if consulKVClient != nil {
		return nil // Sudah diinisialisasi
	}
	conf := consulapi.DefaultConfig()
	if consulAddress != "" {
		conf.Address = consulAddress
	}
	// Tambahkan token jika diperlukan: conf.Token = os.Getenv("CONSUL_HTTP_TOKEN")

	client, err := consulapi.NewClient(conf)
	if err != nil {
		return fmt.Errorf("failed to create consul KV client: %w", err)
	}
	consulKVClient = client
	return nil
}

// getConsulKV mengambil nilai dari Consul KV.
// Mengembalikan nilai dan boolean true jika ditemukan, atau string kosong dan false jika tidak.
func getConsulKV(key string) (string, bool) {
	if consulKVClient == nil {
		// Coba inisialisasi jika belum. Ini hanya akan terjadi jika Load dipanggil sebelum
		// CONSUL_ADDRESS diset atau jika tidak ada service discovery.
		// Idealnya, initConsulKVClient dipanggil di awal Load().
		addr := getEnvString("CONSUL_ADDRESS", "http://localhost:8500")
		if err := initConsulKVClient(addr); err != nil {
			fmt.Printf("Warning: Consul KV client not initialized for key '%s': %v\n", key, err)
			return "", false
		}
	}

	kvPair, _, err := consulKVClient.KV().Get(key, nil)
	if err != nil {
		fmt.Printf("Warning: Error fetching key '%s' from Consul KV: %v\n", key, err)
		return "", false
	}
	if kvPair == nil || kvPair.Value == nil {
		return "", false // Key tidak ditemukan
	}
	return string(kvPair.Value), true
}

// getConfigString membaca string dari Consul KV, lalu env, lalu default.
// consulKey: path di Consul KV (e.g., "config/myapp/loglevel")
// envKey: nama environment variable (e.g., "LOG_LEVEL")
func getConfigString(consulKey, envKey, defaultValue string) string {
	// Coba dari Consul KV
	if consulKey != "" {
		if val, found := getConsulKV(consulKey); found {
			fmt.Printf("Loaded '%s' from Consul KV: %s\n", consulKey, val)
			return val
		}
	}
	// Fallback ke Environment Variable
	if envVal := os.Getenv(envKey); envVal != "" {
		fmt.Printf("Loaded '%s' from ENV: %s\n", envKey, envVal)
		return envVal
	}
	// Fallback ke Default
	fmt.Printf("Using default for '%s'/'%s': %s\n", consulKey, envKey, defaultValue)
	return defaultValue
}

// getConfigInt membaca int dari Consul KV, lalu env, lalu default.
func getConfigInt(consulKey, envKey string, defaultValue int) int {
	strVal := getConfigString(consulKey, envKey, "")
	if strVal != "" {
		if intValue, err := strconv.Atoi(strVal); err == nil {
			return intValue
		}
		fmt.Printf("Warning: Could not parse int from '%s' (value: %s) for Consul key '%s'/env key '%s'. Using default.\n", strVal, consulKey, envKey)
	}
	// Jika dari Consul/Env tidak valid atau kosong, gunakan default
	if envVal := os.Getenv(envKey); envVal != "" { // Cek env lagi jika consul tidak ada/valid
		if intValue, err := strconv.Atoi(envVal); err == nil {
			return intValue
		}
	}
	return defaultValue
}

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
	// Inisialisasi Consul Client untuk KV di awal
	// Alamat Consul bisa dari env atau hardcoded default untuk dev
	consulAddrFromEnv := getEnvString("CONSUL_ADDRESS", "http://localhost:8500")
	if err := initConsulKVClient(consulAddrFromEnv); err != nil {
		fmt.Printf("Warning: Failed to initialize Consul KV client: %v. Configuration might not be loaded from Consul.\n", err)
		// Anda bisa memilih untuk gagal di sini atau melanjutkan dengan env/default
	}

	// Muat JWT Secret (Contoh penggunaan getConfigString)
	// Path di Vault (jika digunakan) vs Path di Consul KV untuk config biasa
	jwtSecret := getConfigString(
		getEnvString("JWT_CONSUL_PATH", ""), // Misal: "config/jwt/secret"
		"JWT_SECRET",
		"your-secret-key-default-fallback",
	)
	// Jika Anda masih menggunakan Vault untuk JWT_SECRET, logic Vault di sini akan diutamakan
	// Atau buat fungsi getSecret yang lebih canggih.

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
			Secret:         jwtSecret, // Sudah dimuat di atas
			ExpirationTime: getConfigInt("config/jwt/expiration", "JWT_EXPIRATION", DefaultJWTExpiration),
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
