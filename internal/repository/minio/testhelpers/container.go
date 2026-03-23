// Package testhelpers предоставляет вспомогательные утилиты для интеграционных тестов MinIO.
package testhelpers

import (
	"context"
	"fmt"

	miniogo "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/testcontainers/testcontainers-go/modules/minio"
)

const (
	accessKey = "minioadmin"
	secretKey = "minioadmin"
)

// MinioContainer представляет тестовый контейнер MinIO.
type MinioContainer struct {
	*minio.MinioContainer
	Endpoint string
}

// NewMinioContainer создаёт и запускает контейнер MinIO для тестирования.
func NewMinioContainer(ctx context.Context) (*MinioContainer, error) {
	container, err := minio.Run(ctx,
		"minio/minio:RELEASE.2024-01-16T16-07-38Z",
		minio.WithUsername(accessKey),
		minio.WithPassword(secretKey),
	)
	if err != nil {
		return nil, fmt.Errorf("запуск контейнера minio: %w", err)
	}

	endpoint, err := container.ConnectionString(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("получение endpoint minio: %w", err)
	}

	return &MinioContainer{
		MinioContainer: container,
		Endpoint:       endpoint,
	}, nil
}

// Terminate останавливает и удаляет контейнер.
func (c *MinioContainer) Terminate(ctx context.Context) error {
	return c.MinioContainer.Terminate(ctx)
}

// NewClient создаёт клиент MinIO, подключённый к тестовому контейнеру.
func (c *MinioContainer) NewClient() (*miniogo.Client, error) {
	client, err := miniogo.New(c.Endpoint, &miniogo.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("создание клиента minio: %w", err)
	}
	return client, nil
}
