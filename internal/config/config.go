package config

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
)

// ServerConfig конфигурация сервера.
type ServerConfig struct {
	// Address адрес HTTP-сервера.
	Address string `env:"ADDRESS"`
	// DatabaseURI URI подключения к PostgreSQL.
	DatabaseURI string `env:"DATABASE_URI"`
	// MinioEndpoint адрес MinIO-сервера.
	MinioEndpoint string `env:"MINIO_ENDPOINT"`
	// MinioAccessKey ключ доступа MinIO.
	MinioAccessKey string `env:"MINIO_ACCESS_KEY"`
	// MinioSecretKey секретный ключ MinIO.
	MinioSecretKey string `env:"MINIO_SECRET_KEY"`
	// MinioSecure использовать TLS при подключении к MinIO.
	MinioSecure bool `env:"MINIO_SECURE"`
	// JWTSecret секрет для подписи JWT-токенов.
	JWTSecret string `env:"JWT_SECRET"`
	// EncryptionKey ключ шифрования AES-256 (32 байта).
	EncryptionKey string `env:"ENCRYPTION_KEY"`
	// MaxBinarySize максимальный размер бинарного файла в байтах.
	MaxBinarySize int64 `env:"MAX_BINARY_SIZE_BYTES"`
}

// LoadServerConfig загружает конфигурацию из флагов командной строки и переменных окружения.
// Переменные окружения имеют приоритет.
func LoadServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{}
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	flagSet.StringVar(&cfg.Address, "a", "localhost:8080", "HTTP server address")
	flagSet.StringVar(&cfg.DatabaseURI, "d", "postgres://keeper:keeperpassword@localhost:5432/gophkeeper?sslmode=disable", "PostgreSQL database URI")
	flagSet.StringVar(&cfg.MinioEndpoint, "m", "localhost:9000", "Minio endpoint")
	flagSet.StringVar(&cfg.MinioAccessKey, "minio-access", "minioadmin", "Minio access key")
	flagSet.StringVar(&cfg.MinioSecretKey, "minio-secret", "minioadmin", "Minio secret key")
	flagSet.BoolVar(&cfg.MinioSecure, "minio-secure", false, "Use TLS when connecting to Minio")
	flagSet.StringVar(&cfg.JWTSecret, "jwt-secret", "", "JWT secret key")
	flagSet.StringVar(&cfg.EncryptionKey, "enc-key", "", "Encryption key (32 bytes)")
	flagSet.Int64Var(&cfg.MaxBinarySize, "max-binary-size", 10<<20, "Maximum binary payload size in bytes")

	if err := flagSet.Parse(filterTestingArgs(os.Args[1:])); err != nil {
		return nil, fmt.Errorf("failed to parse command-line flags: %w", err)
	}

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("jwt secret is required")
	}
	if cfg.EncryptionKey == "" {
		return nil, fmt.Errorf("encryption key is required")
	}
	if len(cfg.EncryptionKey) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes long")
	}
	if cfg.MaxBinarySize <= 0 {
		return nil, fmt.Errorf("max binary size must be greater than zero")
	}

	return cfg, nil
}

func filterTestingArgs(args []string) []string {
	filtered := make([]string, 0, len(args))
	skipNext := false

	for _, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if strings.HasPrefix(arg, "-test.") {
			if !strings.Contains(arg, "=") {
				skipNext = true
			}
			continue
		}
		filtered = append(filtered, arg)
	}

	return filtered
}
