package vault

import (
	"fmt"
	"os"
	"time"

	vault "github.com/hashicorp/vault/api"
)

type Client struct {
	client    *vault.Client
	mountPath string
	basePath  string
}

// NewClient creates a new Vault client
func NewClient() (*Client, error) {
	config := vault.DefaultConfig()
	config.Address = os.Getenv("VAULT_ADDR")

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	client.SetToken(os.Getenv("VAULT_TOKEN"))

	return &Client{
		client:    client,
		mountPath: getEnvOrDefault("VAULT_MOUNT_PATH", "secret"),
		basePath:  getEnvOrDefault("VAULT_BASE_PATH", "xz-platform/users/mnemonics"),
	}, nil
}

// StoreMnemonic stores mnemonic in Vault
// Path: secret/data/xz-platform/users/mnemonics/{did}
func (c *Client) StoreMnemonic(did, mnemonic string) error {
	// 构建路径: secret/data/xz-platform/users/mnemonics/{did}
	path := fmt.Sprintf("%s/data/%s/%s", c.mountPath, c.basePath, did)

	// 准备数据
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"mnemonic":   mnemonic,
			"created_at": time.Now().UTC().Format(time.RFC3339),
			"version":    "1",
		},
	}

	// 写入 Vault
	_, err := c.client.Logical().Write(path, data)
	if err != nil {
		return fmt.Errorf("failed to store mnemonic: %w", err)
	}

	return nil
}

// GetMnemonic retrieves mnemonic from Vault
func (c *Client) GetMnemonic(did string) (string, error) {
	// 构建路径
	path := fmt.Sprintf("%s/data/%s/%s", c.mountPath, c.basePath, did)

	// 读取 Vault
	secret, err := c.client.Logical().Read(path)
	if err != nil {
		return "", fmt.Errorf("failed to read mnemonic: %w", err)
	}

	if secret == nil || secret.Data == nil {
		return "", fmt.Errorf("mnemonic not found for DID: %s", did)
	}

	// 解析数据 (KV v2 格式)
	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid vault data format")
	}

	mnemonic, ok := data["mnemonic"].(string)
	if !ok {
		return "", fmt.Errorf("mnemonic field not found")
	}

	return mnemonic, nil
}

// DeleteMnemonic deletes mnemonic from Vault (for cleanup)
func (c *Client) DeleteMnemonic(did string) error {
	path := fmt.Sprintf("%s/data/%s/%s", c.mountPath, c.basePath, did)

	_, err := c.client.Logical().Delete(path)
	if err != nil {
		return fmt.Errorf("failed to delete mnemonic: %w", err)
	}

	return nil
}

// CheckConnection tests Vault connection
func (c *Client) CheckConnection() error {
	_, err := c.client.Sys().Health()
	if err != nil {
		return fmt.Errorf("vault connection failed: %w", err)
	}
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
