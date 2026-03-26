package ports

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
)

//go:generate mockgen -source=minio.go -destination=mocks/minio_mock.go -package=mocks

// MinioClient определяет минимально необходимые операции для работы с MinIO.
type MinioClient interface {
	// BucketExists проверяет существование бакета.
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	// MakeBucket создает новый бакет.
	MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error
	// PutObject загружает объект в бакет.
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	// GetObject возвращает объект из бакета.
	GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error)
	// RemoveObject удаляет объект из бакета.
	RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error
}
