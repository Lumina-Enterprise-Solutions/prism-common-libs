package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/secrets" // Import Vault client Anda
	consulapi "github.com/hashicorp/consul/api"
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

// Helper untuk mengambil nilai dari map konfigurasi yang dimuat dari Vault, dengan fallback
func getStringFromMap(configMap map[string]string, key string, defaultValue string) string {
	if val, ok := configMap[key]; ok {
		return val
	}
	fmt.Printf("Warning: Key '%s' not found in Vault config map. Using default: '%s'\n", key, defaultValue)
	return defaultValue
}

func getIntFromMap(configMap map[string]string, key string, defaultValue int) int {
	if strVal, ok := configMap[key]; ok {
		if intVal, err := strconv.Atoi(strVal); err == nil {
			return intVal
		}
		fmt.Printf("Warning: Failed to parse int for key '%s' (value: '%s') from Vault config map. Using default: %d\n", key, strVal, defaultValue)
	} else {
		fmt.Printf("Warning: Key '%s' not found in Vault config map. Using default: %d\n", key, defaultValue)
	}
	return defaultValue
}

func Load() (*Config, error) {
	// 1. Dapatkan VAULT_ADDR dan VAULT_TOKEN dari environment (ini adalah bootstrap vars)
	vaultAddr := os.Getenv("VAULT_ADDR")
	vaultToken := os.Getenv("VAULT_TOKEN") // Atau mekanisme auth lain

	if vaultAddr == "" {
		return nil, fmt.Errorf("VAULT_ADDR environment variable not set")
	}
	// Untuk dev, kita bisa asumsikan VAULT_TOKEN ada. Di prod, gunakan AppRole atau auth method lain.
	if vaultToken == "" && os.Getenv("ENVIRONMENT") == "development" { // Cek jika ENVIRONMENT diset
		// Ini asumsi, lebih baik VAULT_TOKEN diset eksplisit
		fmt.Println("Warning: VAULT_TOKEN not set, assuming dev mode (roottoken may or may not work depending on SDK).")
	}

	// 2. Inisialisasi Vault Client
	// NewVaultClient() sudah diubah untuk membaca VAULT_ADDR dan VAULT_TOKEN dari env
	vClient, err := secrets.NewVaultClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Vault client: %w", err)
	}

	// 3. Tentukan path konfigurasi di Vault
	// Ini bisa juga dari env var jika Anda ingin lebih fleksibel
	vaultConfigPath := os.Getenv("VAULT_CONFIG_PATH")
	if vaultConfigPath == "" {
		vaultConfigPath = "config/prism-auth-service" // Default path
		fmt.Printf("VAULT_CONFIG_PATH not set, using default: %s\n", vaultConfigPath)
	}

	// 4. Baca semua konfigurasi dari Vault
	configMap, err := vClient.ReadAllSecrets(vaultConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration from Vault path '%s': %w", vaultConfigPath, err)
	}
	fmt.Printf("Successfully loaded configuration map from Vault path '%s'.\n", vaultConfigPath)

	// 5. Isi struct Config menggunakan nilai dari configMap
	cfg := &Config{
		Environment: getStringFromMap(configMap, "environment", "development"),
		TenantID:    getStringFromMap(configMap, "tenant_id", "default"),
		Consul: ConsulConfig{
			Address: getStringFromMap(configMap, "consul_address", "http://localhost:8500"),
			Token:   getStringFromMap(configMap, "consul_token", ""),
		},
		Database: DatabaseConfig{
			Host:     getStringFromMap(configMap, "db_host", "localhost"),
			Port:     getIntFromMap(configMap, "db_port", DefaultDBPort),
			Database: getStringFromMap(configMap, "db_name", "prism_erp"),
			Username: getStringFromMap(configMap, "db_user", "prism"),
			Password: getStringFromMap(configMap, "db_password", ""), // Default kosong, harus ada di Vault
			SSLMode:  getStringFromMap(configMap, "db_ssl_mode", "disable"),
		},
		Redis: RedisConfig{
			Host:     getStringFromMap(configMap, "redis_host", "localhost"),
			Port:     getIntFromMap(configMap, "redis_port", DefaultRedisPort),
			Password: getStringFromMap(configMap, "redis_password", ""), // Default kosong, harus ada di Vault
			DB:       getIntFromMap(configMap, "redis_db", 0),
		},
		JWT: JWTConfig{
			Secret:         getStringFromMap(configMap, "jwt_secret", ""), // Default kosong, harus ada di Vault
			ExpirationTime: getIntFromMap(configMap, "jwt_expiration", DefaultJWTExpiration),
		},
		Server: ServerConfig{
			Port:         getIntFromMap(configMap, "server_port", DefaultServerPort),
			ReadTimeout:  getIntFromMap(configMap, "server_read_timeout", DefaultReadTimeout),
			WriteTimeout: getIntFromMap(configMap, "server_write_timeout", DefaultWriteTimeout),
		},
		// Anda bisa menambahkan field lain seperti LogLevel, GinMode ke struct Config jika perlu
		// dan memuatnya dari configMap juga.
	}

	// Inisialisasi Consul Client untuk Service Discovery (jika masih digunakan)
	// Ini terpisah dari pembacaan config dari Vault KV, kecuali jika consul_address dari Vault
	// Jika Anda masih menggunakan Consul untuk service discovery (bukan config KV)
	// Anda perlu initConsulKVClient(cfg.Consul.Address) di sini atau di main.go
	// seperti sebelumnya, karena initConsulKVClient yang lama mungkin tidak dipanggil.
	// Atau, service discovery bisa menggunakan cfg.Consul.Address yang sudah dimuat.
	// Jika Anda membuat `discovery.NewConsulClient(cfg)` di main.go, pastikan cfg.Consul.Address benar.

	return cfg, nil
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.Username, d.Password, d.Database, d.SSLMode)
}

func (r RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// Fungsi getEnvString LAMA (hanya baca dari env, untuk variabel yang tidak perlu Consul KV)
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Fungsi getEnvInt LAMA (hanya baca dari env)
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
