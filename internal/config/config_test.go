package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadServerConfig(t *testing.T) {
	os.Setenv("ADDRESS", "127.0.0.1:9090")
	os.Setenv("DATABASE_URI", "postgres://test")
	os.Setenv("JWT_SECRET", "test-jwt-secret")
	os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")
	defer os.Clearenv()

	cfg, err := LoadServerConfig()
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1:9090", cfg.Address)
	assert.Equal(t, "postgres://test", cfg.DatabaseURI)
}

func TestLoadServerConfig_MissingSecrets(t *testing.T) {
	defer os.Clearenv()

	cfg, err := LoadServerConfig()
	assert.Nil(t, cfg)
	assert.Error(t, err)
}
