package client

import (
	"fmt"
	"log"
	"os"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// VaultClient adalah wrapper untuk klien Vault resmi.
type VaultClient struct {
	client *vaultapi.Client
}

// NewVaultClient membuat dan mengonfigurasi klien baru untuk berinteraksi dengan Vault.
func NewVaultClient() (*VaultClient, error) {
	// Konfigurasi ini membaca alamat Vault dan token dari environment variables,
	// yang akan kita set di docker-compose.yml.
	config := vaultapi.DefaultConfig()

	// Alamat Vault di dalam jaringan Docker
	vaultAddr := os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		vaultAddr = "http://vault:8200" // Default jika tidak diset
	}
	config.Address = vaultAddr

	client, err := vaultapi.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat klien vault: %w", err)
	}

	// Gunakan root token yang kita definisikan di docker-compose
	vaultToken := os.Getenv("VAULT_TOKEN")
	if vaultToken == "" {
		vaultToken = "root-token-for-dev" // Default untuk dev
	}
	client.SetToken(vaultToken)

	return &VaultClient{client: client}, nil
}

// ReadSecret mengambil sebuah rahasia dari Vault.
// path: path ke secret engine, misal "secret/data/prism"
// key: nama field di dalam rahasia, misal "jwt_secret"
func (vc *VaultClient) ReadSecret(path, key string) (string, error) {
	// UBAH: Gunakan client.Logical().Read() yang lebih umum.
	// Ini adalah cara yang lebih standar untuk membaca dari secret engine manapun.
	// Untuk KV v2, path harus menyertakan "data", contoh: "secret/data/prism".
	secret, err := vc.client.Logical().Read(path)
	if err != nil {
		return "", fmt.Errorf("gagal membaca rahasia dari path '%s': %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return "", fmt.Errorf("tidak ada rahasia yang ditemukan di path '%s'", path)
	}

	// Untuk KV v2, data rahasia yang sebenarnya ada di dalam map bernama 'data'.
	// Jadi kita perlu mengakses secret.Data["data"]
	secretData, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("format data rahasia tidak valid di path '%s'", path)
	}

	// Sekarang ambil key spesifik dari dalam map 'data' tersebut.
	value, ok := secretData[key]
	if !ok {
		return "", fmt.Errorf("key '%s' tidak ditemukan di dalam data rahasia di path '%s'", key, path)
	}

	// Konversi nilai (yang tipenya interface{}) ke string.
	valueStr, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("nilai rahasia untuk key '%s' bukan string", key)
	}

	return valueStr, nil
}
func (vc *VaultClient) ReadMultipleSecrets(path string, keys ...string) (map[string]string, error) {
	secret, err := vc.client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca rahasia dari path '%s': %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("tidak ada rahasia yang ditemukan di path '%s'", path)
	}

	secretData, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("format data rahasia tidak valid di path '%s'", path)
	}

	results := make(map[string]string)
	for _, key := range keys {
		value, ok := secretData[key]
		if !ok {
			return nil, fmt.Errorf("key '%s' tidak ditemukan di dalam data rahasia di path '%s'", key, path)
		}
		valueStr, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("nilai rahasia untuk key '%s' bukan string", key)
		}
		results[key] = valueStr
	}

	return results, nil
}
func (vc *VaultClient) LoadSecretsToEnv(path string, keys ...string) error {
	secretsMap, err := vc.ReadMultipleSecrets(path, keys...)
	if err != nil {
		return err
	}

	for key, secretValue := range secretsMap {
		if err := os.Setenv(strings.ToUpper(key), secretValue); err != nil {
			return fmt.Errorf("gagal mengatur env var untuk '%s': %w", key, err)
		}
		log.Printf("Berhasil memuat rahasia '%s' dari Vault dan mengaturnya sebagai env var.", key)
	}
	return nil
}
