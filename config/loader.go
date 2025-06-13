// file: common/prism-common-libs/config/loader.go
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	consulapi "github.com/hashicorp/consul/api"
)

type Loader struct {
	client *consulapi.Client
}

func NewLoader() (*Loader, error) {
	config := consulapi.DefaultConfig()
	consulAddr := os.Getenv("CONSUL_ADDR")
	if consulAddr == "" {
		consulAddr = "http://consul:8500"
	}
	config.Address = consulAddr

	client, err := consulapi.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat klien consul: %w", err)
	}

	return &Loader{client: client}, nil
}

// Get retrieves a config value. It first checks environment variables,
// then falls back to Consul KV.
func (l *Loader) Get(key string, defaultValue string) string {
	// Prioritas 1: Environment Variable (memungkinkan override saat runtime)
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	// Prioritas 2: Consul KV Store
	kvPair, _, err := l.client.KV().Get(key, nil)
	if err != nil {
		log.Printf("Peringatan: Gagal membaca key '%s' dari Consul: %v. Menggunakan default.", key, err)
		return defaultValue
	}
	if kvPair != nil {
		return string(kvPair.Value)
	}

	// Prioritas 3: Nilai Default
	return defaultValue
}

// GetInt adalah helper untuk mengambil nilai integer.
func (l *Loader) GetInt(key string, defaultValue int) int {
	valStr := l.Get(key, "")
	if valStr == "" {
		return defaultValue
	}
	valInt, err := strconv.Atoi(valStr)
	if err != nil {
		log.Printf("Peringatan: Gagal konversi key '%s' ke int: %v. Menggunakan default.", key, err)
		return defaultValue
	}
	return valInt
}
