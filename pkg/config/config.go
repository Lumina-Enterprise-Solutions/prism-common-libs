// File: prism-common-libs/pkg/config/config.go
package config

import (
	"context" // <-- TAMBAHKAN IMPORT INI
	"fmt"
	"os"
	"strconv"

	// "time" // Tambahkan jika Anda ingin menggunakan context.WithTimeout

	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/secrets" // Import Vault client Anda
	// consulapi "github.com/hashicorp/consul/api" // Tidak dibutuhkan jika tidak ada lagi Consul KV read di sini
)

// Hapus bagian Consul KV jika tidak lagi digunakan untuk memuat konfigurasi:
// var consulKVClient *consulapi.Client
// func initConsulKVClient(...) { ... }
// func getConsulKV(...) (string, bool) { ... }
// func getConfigString(...) string { ... } // Ini digantikan oleh getStringFromMap
// func getConfigInt(...) int { ... }     // Ini digantikan oleh getIntFromMap

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
	ReadTimeout  int // Dalam detik
	WriteTimeout int // Dalam detik
	// GinMode      string // Opsional: bisa ditambahkan di sini jika ingin dari Vault
}

type JWTConfig struct {
	Secret         string
	ExpirationTime int // Dalam detik
}

type ConsulConfig struct { // Ini untuk konfigurasi koneksi ke Consul untuk Service Discovery
	Address string
	Token   string
}

type LDAPConfig struct {
	Host         string `mapstructure:"host"`          // e.g., "ad.example.com"
	Port         int    `mapstructure:"port"`          // e.g., 389 (LDAP) atau 636 (LDAPS)
	UseTLS       bool   `mapstructure:"use_tls"`       // Apakah menggunakan StartTLS atau LDAPS
	BindDN       string `mapstructure:"bind_dn"`       // DN akun layanan untuk binding/pencarian, e.g., "CN=svc_ldap,OU=ServiceAccounts,DC=example,DC=com"
	BindPassword string `mapstructure:"bind_password"` // Password akun layanan (dari Vault!)
	UserBaseDN   string `mapstructure:"user_base_dn"`  // Base DN untuk mencari user, e.g., "OU=Users,DC=example,DC=com"
	GroupBaseDN  string `mapstructure:"group_base_dn"` // Base DN untuk mencari grup (jika perlu)
	// Atribut AD yang akan dipetakan ke field User lokal
	ADAttributeEmail      string `mapstructure:"ad_attribute_email"`       // e.g., "mail"
	ADAttributeFirstName  string `mapstructure:"ad_attribute_first_name"`  // e.g., "givenName"
	ADAttributeLastName   string `mapstructure:"ad_attribute_last_name"`   // e.g., "sn"
	ADAttributeUPN        string `mapstructure:"ad_attribute_upn"`         // e.g., "userPrincipalName"
	ADAttributeObjectGUID string `mapstructure:"ad_attribute_object_guid"` // e.g., "objectGUID" (perlu konversi dari binary)
	ADAttributeMemberOf   string `mapstructure:"ad_attribute_member_of"`   // e.g., "memberOf"
}

type Config struct {
	Environment string
	TenantID    string
	ServiceName string // Harus diisi dari Vault
	// LogLevel    string // Opsional: jika ingin level log dari Vault
	LDAP LDAPConfig // <-- Pastikan ini ada

	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Server   ServerConfig
	Consul   ConsulConfig // Info koneksi ke Consul untuk Service Discovery
}

const (
	DefaultDBPort        = 5432
	DefaultRedisPort     = 6379
	DefaultJWTExpiration = 3600 // 1 jam
	DefaultServerPort    = 8080
	DefaultReadTimeout   = 10 // detik
	DefaultWriteTimeout  = 10 // detik
)

// Helper untuk mengambil nilai string dari map konfigurasi, dengan fallback
func getStringFromMap(configMap map[string]string, key string, defaultValue string) string {
	if val, ok := configMap[key]; ok && val != "" { // Tambahkan cek val != ""
		return val
	}
	// Tidak perlu print warning di sini jika default digunakan, kecuali jika key wajib ada
	// fmt.Printf("Warning: Key '%s' not found or empty in Vault config map. Using default: '%s'\n", key, defaultValue)
	return defaultValue
}

// Helper untuk mengambil nilai integer dari map konfigurasi, dengan fallback
func getIntFromMap(configMap map[string]string, key string, defaultValue int) int {
	if strVal, ok := configMap[key]; ok && strVal != "" { // Tambahkan cek strVal != ""
		if intVal, err := strconv.Atoi(strVal); err == nil {
			return intVal
		}
		fmt.Printf("Warning: Failed to parse int for key '%s' (value: '%s') from Vault config map. Using default: %d\n", key, strVal, defaultValue)
	} else {
		// fmt.Printf("Warning: Key '%s' not found or empty in Vault config map. Using default: %d\n", key, defaultValue)
	}
	return defaultValue
}

func Load() (*Config, error) {
	// 1. Buat parent context untuk operasi load konfigurasi
	// Anda bisa menambahkan timeout jika perlu, contoh:
	// loadCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 10 detik timeout
	// defer cancel()
	loadCtx := context.Background() // Untuk sekarang, cukup dengan Background context

	// 2. Dapatkan VAULT_ADDR dan VAULT_TOKEN dari environment (bootstrap variables)
	vaultAddr := os.Getenv("VAULT_ADDR")
	// VAULT_TOKEN akan otomatis dibaca oleh NewVaultClient jika diset di env.
	// Jika Anda perlu logika khusus untuk VAULT_TOKEN (misalnya, membaca dari file jika tidak di env),
	// Anda bisa melakukannya di sini sebelum memanggil NewVaultClient atau di dalam NewVaultClient.

	if vaultAddr == "" {
		return nil, fmt.Errorf("VAULT_ADDR environment variable is not set. Cannot connect to Vault")
	}
	// Logika untuk VAULT_TOKEN di development (jika perlu):
	// if os.Getenv("VAULT_TOKEN") == "" && (os.Getenv("ENVIRONMENT") == "development" || os.Getenv("GIN_MODE") == "debug") {
	//     fmt.Println("Warning: VAULT_TOKEN environment variable is not set. Vault client might fail if not running in dev mode with a known root token or if auth method is not configured.")
	// }

	// 3. Inisialisasi Vault Client
	vClient, err := secrets.NewVaultClient() // NewVaultClient sudah diperbarui untuk menggunakan env vars
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Vault client: %w", err)
	}

	// 4. Tentukan path konfigurasi di Vault
	vaultConfigPath := os.Getenv("VAULT_CONFIG_PATH")
	if vaultConfigPath == "" {
		vaultConfigPath = "config/prism-auth-service" // Default path ke secret
		fmt.Printf("VAULT_CONFIG_PATH environment variable not set, using default: '%s'\n", vaultConfigPath)
	}

	// Tentukan mount path KV engine Anda.
	kvMountPath := os.Getenv("VAULT_KV_MOUNT_PATH")
	if kvMountPath == "" {
		kvMountPath = "secret" // Default mount path untuk KV v2
		fmt.Printf("VAULT_KV_MOUNT_PATH environment variable not set, using default: '%s'\n", kvMountPath)
	}

	// 5. Baca semua konfigurasi dari Vault, teruskan context dan mount path
	fmt.Printf("Attempting to load configuration from Vault: Mount='%s', Path='%s'\n", kvMountPath, vaultConfigPath)
	configMap, err := vClient.ReadAllSecrets(loadCtx, kvMountPath, vaultConfigPath) // <-- PERUBAHAN DI SINI
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration from Vault (mount: %s, path: %s): %w", kvMountPath, vaultConfigPath, err)
	}
	fmt.Printf("Successfully loaded %d configuration entries from Vault path '%s'.\n", len(configMap), vaultConfigPath)

	// 6. Isi struct Config menggunakan nilai dari configMap
	cfg := &Config{
		Environment: getStringFromMap(configMap, "environment", "development"),
		TenantID:    getStringFromMap(configMap, "tenant_id", "default"),
		ServiceName: getStringFromMap(configMap, "service_name", ""), // Wajib ada di Vault
		// LogLevel:    getStringFromMap(configMap, "log_level", "info"), // Jika Anda menambahkannya

		Consul: ConsulConfig{ // Info untuk koneksi ke Consul untuk Service Discovery
			Address: getStringFromMap(configMap, "consul_address", "http://localhost:8500"),
			Token:   getStringFromMap(configMap, "consul_token", ""),
		},
		Database: DatabaseConfig{
			Host:     getStringFromMap(configMap, "db_host", "localhost"),
			Port:     getIntFromMap(configMap, "db_port", DefaultDBPort),
			Database: getStringFromMap(configMap, "db_name", "prism_erp"),
			Username: getStringFromMap(configMap, "db_user", "prism"),
			Password: getStringFromMap(configMap, "db_password", ""), // Wajib ada di Vault
			SSLMode:  getStringFromMap(configMap, "db_ssl_mode", "disable"),
		},
		Redis: RedisConfig{
			Host:     getStringFromMap(configMap, "redis_host", "localhost"),
			Port:     getIntFromMap(configMap, "redis_port", DefaultRedisPort),
			Password: getStringFromMap(configMap, "redis_password", ""), // Bisa kosong jika Redis tanpa auth
			DB:       getIntFromMap(configMap, "redis_db", 0),
		},
		JWT: JWTConfig{
			Secret:         getStringFromMap(configMap, "jwt_secret", ""), // Wajib ada di Vault
			ExpirationTime: getIntFromMap(configMap, "jwt_expiration", DefaultJWTExpiration),
		},
		Server: ServerConfig{
			Port:         getIntFromMap(configMap, "server_port", DefaultServerPort), // Wajib ada di Vault
			ReadTimeout:  getIntFromMap(configMap, "server_read_timeout", DefaultReadTimeout),
			WriteTimeout: getIntFromMap(configMap, "server_write_timeout", DefaultWriteTimeout),
			// GinMode:      getStringFromMap(configMap, "gin_mode", "debug"), // Jika Anda menambahkannya
		},
		LDAP: LDAPConfig{
			Host:                  getStringFromMap(configMap, "ldap_host", ""),
			Port:                  getIntFromMap(configMap, "ldap_port", 389),
			UseTLS:                getBoolFromMap(configMap, "ldap_use_tls", false), // Anda perlu helper getBoolFromMap
			BindDN:                getStringFromMap(configMap, "ldap_bind_dn", ""),
			BindPassword:          getStringFromMap(configMap, "ldap_bind_password", ""), // Ini harusnya secret!
			UserBaseDN:            getStringFromMap(configMap, "ldap_user_base_dn", ""),
			GroupBaseDN:           getStringFromMap(configMap, "ldap_group_base_dn", ""),
			ADAttributeEmail:      getStringFromMap(configMap, "ldap_ad_attribute_email", "mail"),
			ADAttributeFirstName:  getStringFromMap(configMap, "ldap_ad_attribute_first_name", "givenName"),
			ADAttributeLastName:   getStringFromMap(configMap, "ldap_ad_attribute_last_name", "sn"),
			ADAttributeUPN:        getStringFromMap(configMap, "ldap_ad_attribute_upn", "userPrincipalName"),
			ADAttributeObjectGUID: getStringFromMap(configMap, "ldap_ad_attribute_object_guid", "objectGUID"),
			ADAttributeMemberOf:   getStringFromMap(configMap, "ldap_ad_attribute_member_of", "memberOf"),
		},
	}

	// Validasi konfigurasi penting (contoh)
	if cfg.ServiceName == "" {
		return nil, fmt.Errorf("critical configuration 'service_name' is missing from Vault")
	}
	if cfg.Server.Port == 0 {
		return nil, fmt.Errorf("critical configuration 'server_port' is missing from Vault")
	}
	if cfg.Database.Password == "" { // Atau field lain yang wajib
		fmt.Println("Warning: 'db_password' is empty. Ensure this is intentional (e.g., local dev without password).")
	}
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("critical configuration 'jwt_secret' is missing from Vault")
	}

	if cfg.LDAP.Host != "" && cfg.LDAP.BindPassword == "" { // Contoh validasi sederhana
		// Password bisa kosong jika anonymous bind, tapi jarang untuk operasi write/sync
		fmt.Println("Warning: LDAP host is configured but ldap_bind_password is empty.")
	}

	return cfg, nil
}

// Metode DSN dan Address tetap berguna
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.Username, d.Password, d.Database, d.SSLMode)
}

func (r RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// Helper getBoolFromMap (tambahkan jika belum ada)
func getBoolFromMap(configMap map[string]string, key string, defaultValue bool) bool {
	if strVal, ok := configMap[key]; ok && strVal != "" {
		if boolVal, err := strconv.ParseBool(strVal); err == nil {
			return boolVal
		}
		fmt.Printf("Warning: Failed to parse bool for key '%s' (value: '%s') from Vault config map. Using default: %t\n", key, strVal, defaultValue)
	}
	return defaultValue
}
