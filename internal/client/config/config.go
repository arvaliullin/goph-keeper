package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// ClientConfig хранит конфигурацию клиента.
type ClientConfig struct {
	// Token JWT-токен аутентификации.
	Token string `json:"token"`
	// Server URL сервера GophKeeper.
	Server string `json:"server"`
	// LastSync время последней синхронизации в формате RFC3339.
	LastSync string `json:"last_sync,omitempty"`
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".gophkeeper")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// LoadConfig загружает конфигурацию из файла.
func LoadConfig() (*ClientConfig, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ClientConfig{Server: "http://localhost:8080"}, nil
		}
		return nil, err
	}

	var cfg ClientConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Server == "" {
		cfg.Server = "http://localhost:8080"
	}

	return &cfg, nil
}

// SaveConfig сохраняет конфигурацию в файл.
func SaveConfig(cfg *ClientConfig) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// LastSyncTime возвращает время последней синхронизации из конфигурации.
func (c *ClientConfig) LastSyncTime() (time.Time, error) {
	if c.LastSync == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, c.LastSync)
}

// SetLastSync обновляет время последней синхронизации.
func (c *ClientConfig) SetLastSync(syncTime time.Time) {
	c.LastSync = syncTime.UTC().Format(time.RFC3339)
}
