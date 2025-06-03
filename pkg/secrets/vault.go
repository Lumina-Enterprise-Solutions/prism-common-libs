// File: prism-common-libs/pkg/secrets/vault.go
package secrets

import (
	"context" // <-- TAMBAHKAN IMPORT INI
	"fmt"

	// "os" // VAULT_ADDR dan VAULT_TOKEN sudah otomatis dibaca oleh NewClient

	"github.com/hashicorp/vault/api"
)

type VaultClient struct {
	client *api.Client
}

// NewVaultClient creates a new Vault client.
// It relies on VAULT_ADDR and VAULT_TOKEN environment variables being set,
// as api.DefaultConfig() and api.NewClient() will pick them up.
func NewVaultClient() (*VaultClient, error) {
	// api.DefaultConfig() akan mencoba membaca VAULT_ADDR dari environment.
	// api.NewClient(config) akan mencoba membaca VAULT_TOKEN dari environment.
	config := api.DefaultConfig()

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w. Ensure VAULT_ADDR is set and reachable", err)
	}

	// Tidak perlu SetToken manual jika VAULT_TOKEN sudah diset sebagai env var,
	// karena NewClient akan menggunakannya. Jika VAULT_TOKEN tidak diset dan Anda
	// menjalankan Vault dalam mode dev dengan root token, NewClient mungkin tidak
	// otomatis mengambil token dev tersebut jika VAULT_TOKEN env var kosong.
	// Untuk dev dengan root token dan VAULT_TOKEN env var tidak diset, Anda BISA melakukan ini:
	// if client.Token() == "" && os.Getenv("VAULT_DEV_ROOT_TOKEN_ID") != "" {
	//    client.SetToken(os.Getenv("VAULT_DEV_ROOT_TOKEN_ID"))
	// }
	// Namun, praktik terbaik adalah selalu menyetel VAULT_TOKEN env var, bahkan untuk dev.

	return &VaultClient{client: client}, nil
}

// ReadSecretValue reads a specific key from a Vault KV v2 secret path.
// The 'secretMountPath' is the mount path of your KV v2 secrets engine (e.g., "secret" or "kv").
// The 'secretPath' is the path to the secret within that engine (e.g., "config/myapp").
func (vc *VaultClient) ReadSecretValue(ctx context.Context, secretMountPath, secretPath string, key string) (string, error) {
	if ctx == nil {
		// Ini adalah fallback, tapi idealnya context non-nil selalu disediakan oleh pemanggil.
		// Untuk operasi singkat seperti load config, Background() biasanya cukup.
		// commonLogger.Warn("ReadSecretValue called with nil context, using context.Background()")
		ctx = context.Background()
	}

	// Pastikan mount path tidak kosong, default ke "secret" jika perlu
	if secretMountPath == "" {
		secretMountPath = "secret" // Atau ambil dari konfigurasi jika bisa berubah
	}

	secretData, err := vc.client.KVv2(secretMountPath).Get(ctx, secretPath)
	if err != nil {
		return "", fmt.Errorf("failed to read secret from vault (mount: %s, path: %s, key: %s): %w", secretMountPath, secretPath, key, err)
	}

	if secretData == nil || secretData.Data == nil {
		return "", fmt.Errorf("no data found at secret (mount: %s, path: %s, key: %s)", secretMountPath, secretPath, key)
	}

	value, ok := secretData.Data[key].(string)
	if !ok {
		return "", fmt.Errorf("key '%s' not found or not a string in secret (mount: %s, path: %s)", key, secretMountPath, secretPath)
	}
	return value, nil
}

// ReadAllSecrets reads all key-value pairs from a Vault KV v2 secret path.
// The 'secretMountPath' is the mount path of your KV v2 secrets engine.
// The 'secretPath' is the path to the secret within that engine.
func (vc *VaultClient) ReadAllSecrets(ctx context.Context, secretMountPath, secretPath string) (map[string]string, error) {
	if ctx == nil {
		// commonLogger.Warn("ReadAllSecrets called with nil context, using context.Background()")
		ctx = context.Background()
	}

	if secretMountPath == "" {
		secretMountPath = "secret"
	}

	secretData, err := vc.client.KVv2(secretMountPath).Get(ctx, secretPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read all secrets from vault (mount: %s, path: %s): %w", secretMountPath, secretPath, err)
	}

	if secretData == nil || secretData.Data == nil {
		// Jika path valid tapi tidak ada data (misalnya secret baru dibuat tanpa data),
		// ini bukan error fatal. Kembalikan map kosong.
		fmt.Printf("Info: No data/secret found at Vault (mount: %s, path: %s). Returning empty map.\n", secretMountPath, secretPath)
		return make(map[string]string), nil
	}

	resultMap := make(map[string]string)
	for key, value := range secretData.Data {
		strValue, ok := value.(string)
		if !ok {
			// Log warning dan skip key ini daripada membuat seluruh operasi gagal.
			fmt.Printf("Warning: Value for key '%s' at Vault (mount: %s, path: %s) is not a string. Skipping.\n", key, secretMountPath, secretPath)
			continue
		}
		resultMap[key] = strValue
	}
	return resultMap, nil
}
