// File: prism-common-libs/pkg/secrets/vault.go
package secrets

import (
	"fmt"
	// "os" // VAULT_ADDR dan VAULT_TOKEN sudah otomatis dibaca oleh NewClient

	"github.com/hashicorp/vault/api"
)

type VaultClient struct {
	client *api.Client
}

func NewVaultClient() (*VaultClient, error) {
	config := api.DefaultConfig() // Reads VAULT_ADDR, VAULT_TOKEN from env
	// Jika VAULT_TOKEN tidak ada di env dan Anda menggunakan roottoken untuk dev,
	// Anda mungkin perlu menyetelnya secara manual jika SDK tidak mengambilnya.
	// if os.Getenv("VAULT_TOKEN") == "" && os.Getenv("VAULT_ADDR") != "" {
	//     // Hanya untuk dev, jika VAULT_TOKEN env var kosong
	//     // config.Token = "roottoken" // Ini akan di-override jika VAULT_TOKEN diset
	// }

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}
	// Jika VAULT_TOKEN perlu diset eksplisit setelah NewClient (jarang)
	// client.SetToken(os.Getenv("VAULT_TOKEN"))

	return &VaultClient{client: client}, nil
}

// ReadSecret reads a specific key from a Vault KV v2 secret path.
func (vc *VaultClient) ReadSecretValue(path string, key string) (string, error) {
	secretData, err := vc.client.KVv2("secret").Get(nil, path) // "secret" adalah mount path KV
	if err != nil {
		return "", fmt.Errorf("failed to read secret from vault path %s: %w", path, err)
	}
	if secretData == nil || secretData.Data == nil {
		return "", fmt.Errorf("no data found at secret path %s", path)
	}
	value, ok := secretData.Data[key].(string)
	if !ok {
		return "", fmt.Errorf("key '%s' not found or not a string in secret at path %s", key, path)
	}
	return value, nil
}

// ReadAllSecrets reads all key-value pairs from a Vault KV v2 secret path.
// Returns a map[string]string.
func (vc *VaultClient) ReadAllSecrets(path string) (map[string]string, error) {
	secretData, err := vc.client.KVv2("secret").Get(nil, path) // "secret" adalah mount path KV
	if err != nil {
		return nil, fmt.Errorf("failed to read all secrets from vault path %s: %w", path, err)
	}

	if secretData == nil || secretData.Data == nil {
		return nil, fmt.Errorf("no data found at secret path %s", path)
	}

	resultMap := make(map[string]string)
	for key, value := range secretData.Data {
		strValue, ok := value.(string)
		if !ok {
			// Anda bisa memilih untuk skip non-string atau return error
			fmt.Printf("Warning: Value for key '%s' at path '%s' is not a string, skipping.\n", key, path)
			continue
		}
		resultMap[key] = strValue
	}
	return resultMap, nil
}
