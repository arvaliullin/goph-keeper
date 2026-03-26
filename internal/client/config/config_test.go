package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClientConfig(t *testing.T) {
	// mock home dir
	home := t.TempDir()
	t.Setenv("HOME", home)

	// test save
	cfg := &ClientConfig{
		Token:  "test_token",
		Server: "http://testserver",
	}
	err := SaveConfig(cfg)
	assert.NoError(t, err)

	// test load
	loaded, err := LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "test_token", loaded.Token)
	assert.Equal(t, "http://testserver", loaded.Server)
}

func TestClientConfig_NoFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	loaded, err := LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8080", loaded.Server)
}

func TestClientConfig_LastSync(t *testing.T) {
	cfg := &ClientConfig{}
	now := time.Now().UTC().Truncate(time.Second)
	cfg.SetLastSync(now)

	parsed, err := cfg.LastSyncTime()
	assert.NoError(t, err)
	assert.Equal(t, now.Format(time.RFC3339), parsed.Format(time.RFC3339))
}
