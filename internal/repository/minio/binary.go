package minio

import (
	"bytes"
	"context"
	"io"

	"github.com/arvaliullin/goph-keeper/internal/core/ports"
	retrypkg "github.com/arvaliullin/goph-keeper/internal/pkg/retry"
	"github.com/minio/minio-go/v7"
)

// BinaryRepository реализация хранилища бинарных файлов в Minio.
type BinaryRepository struct {
	client     ports.MinioClient
	bucketName string
	strategy   *retrypkg.Strategy
}

// NewBinaryRepository создает новый экземпляр BinaryRepository.
func NewBinaryRepository(ctx context.Context, client ports.MinioClient, bucketName string) (*BinaryRepository, error) {
	return NewBinaryRepositoryWithStrategy(ctx, client, bucketName, retrypkg.NewStrategy(nil, IsRetryable))
}

// NewBinaryRepositoryWithStrategy создает BinaryRepository с кастомной стратегией retry.
func NewBinaryRepositoryWithStrategy(
	ctx context.Context,
	client ports.MinioClient,
	bucketName string,
	strategy *retrypkg.Strategy,
) (*BinaryRepository, error) {
	if strategy == nil {
		strategy = retrypkg.NewStrategy(nil, IsRetryable)
	}

	var (
		exists bool
		err    error
	)

	err = strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		exists, err = client.BucketExists(ctx, bucketName)
		return err
	})
	if err != nil {
		return nil, err
	}

	if !exists {
		err = strategy.DoWithRetry(ctx, func(ctx context.Context) error {
			makeErr := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
			errResp := minio.ToErrorResponse(makeErr)
			if errResp.Code == minio.BucketAlreadyExists || errResp.Code == minio.BucketAlreadyOwnedByYou {
				return nil
			}
			return makeErr
		})
		if err != nil {
			return nil, err
		}
	}

	return &BinaryRepository{
		client:     client,
		bucketName: bucketName,
		strategy:   strategy,
	}, nil
}

// Upload загружает файл в Minio.
func (r *BinaryRepository) Upload(ctx context.Context, objectName string, reader io.Reader, size int64) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	return r.strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		_, err = r.client.PutObject(
			ctx,
			r.bucketName,
			objectName,
			bytes.NewReader(data),
			int64(len(data)),
			minio.PutObjectOptions{
				ContentType: "application/octet-stream",
			},
		)
		return err
	})
}

// Download скачивает файл из Minio.
func (r *BinaryRepository) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	var (
		object *minio.Object
		err    error
	)

	err = r.strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		object, err = r.client.GetObject(ctx, r.bucketName, objectName, minio.GetObjectOptions{})
		return err
	})
	if err != nil {
		return nil, err
	}
	return object, nil
}

// Delete удаляет файл из Minio.
func (r *BinaryRepository) Delete(ctx context.Context, objectName string) error {
	return r.strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		return r.client.RemoveObject(ctx, r.bucketName, objectName, minio.RemoveObjectOptions{})
	})
}
