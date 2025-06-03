// File: prism-common-libs/pkg/secrets/vault.go
package secrets

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

type VaultClient struct {
	client *api.Client
}

// NewVaultClient creates a new Vault client.
// VAULT_ADDR and VAULT_TOKEN environment variables are typically used by the SDK.
func NewVaultClient() (*VaultClient, error) {
	config := api.DefaultConfig() // Reads VAULT_ADDR from env

	// Jika VAULT_ADDR tidak diset di env, bisa diset manual di sini
	// if os.Getenv("VAULT_ADDR") == "" {
	//    config.Address = "http://localhost:8200" // Atau dari config file
	// }

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	// Jika VAULT_TOKEN tidak diset di env, bisa diset manual di sini
	// token := os.Getenv("VAULT_TOKEN")
	// if token == "" {
	//    token = "roottoken" // HANYA UNTUK DEV
	// }
	// client.SetToken(token) // Token juga bisa dibaca dari env VAULT_TOKEN otomatis

	return &VaultClient{client: client}, nil
}

// ReadSecret reads a secret from Vault KV v2 engine.
func (vc *VaultClient) ReadSecret(path string, key string) (string, error) {
	// Path untuk KV v2 biasanya: <mount_path>/data/<secret_path>
	// Contoh: jika mount path adalah "secret" dan path secret adalah "prism/auth",
	// maka full path adalah "secret/data/prism/auth"
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
