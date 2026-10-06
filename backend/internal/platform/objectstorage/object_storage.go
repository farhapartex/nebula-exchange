package objectstorage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/config"
)

type URLSigner interface {
	PresignedDownloadURL(ctx context.Context, objectKey string) (string, error)
}

type Uploader interface {
	EnsureBucket(ctx context.Context) error
	Upload(ctx context.Context, objectKey string, content io.Reader, size int64, contentType string) error
}

type MinioStorage struct {
	internalClient  *minio.Client
	publicClient    *minio.Client
	bucket          string
	region          string
	presignLifetime time.Duration
}

func NewMinioStorage(storageConfig config.StorageConfig) (*MinioStorage, error) {
	internalClient, err := newClient(storageConfig.Endpoint, storageConfig)
	if err != nil {
		return nil, fmt.Errorf("create storage client: %w", err)
	}
	publicClient, err := newClient(storageConfig.PublicEndpoint, storageConfig)
	if err != nil {
		return nil, fmt.Errorf("create public storage client: %w", err)
	}
	return &MinioStorage{
		internalClient:  internalClient,
		publicClient:    publicClient,
		bucket:          storageConfig.Bucket,
		region:          storageConfig.Region,
		presignLifetime: storageConfig.PresignLifetime,
	}, nil
}

func newClient(endpoint string, storageConfig config.StorageConfig) (*minio.Client, error) {
	return minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(storageConfig.AccessKey, storageConfig.SecretKey, ""),
		Secure: storageConfig.UseSSL,
		Region: storageConfig.Region,
	})
}

func (storage *MinioStorage) PresignedDownloadURL(ctx context.Context, objectKey string) (string, error) {
	presignedURL, err := storage.publicClient.PresignedGetObject(ctx, storage.bucket, objectKey, storage.presignLifetime, nil)
	if err != nil {
		return "", fmt.Errorf("presign %s: %w", objectKey, err)
	}
	return presignedURL.String(), nil
}

func (storage *MinioStorage) EnsureBucket(ctx context.Context) error {
	bucketExists, err := storage.internalClient.BucketExists(ctx, storage.bucket)
	if err != nil {
		return fmt.Errorf("check bucket %s: %w", storage.bucket, err)
	}
	if bucketExists {
		return nil
	}
	if err := storage.internalClient.MakeBucket(ctx, storage.bucket, minio.MakeBucketOptions{Region: storage.region}); err != nil {
		return fmt.Errorf("create bucket %s: %w", storage.bucket, err)
	}
	return nil
}

func (storage *MinioStorage) Upload(ctx context.Context, objectKey string, content io.Reader, size int64, contentType string) error {
	_, err := storage.internalClient.PutObject(ctx, storage.bucket, objectKey, content, size, minio.PutObjectOptions{
		ContentType:  contentType,
		CacheControl: "private, max-age=3600",
	})
	if err != nil {
		return fmt.Errorf("upload %s: %w", objectKey, err)
	}
	return nil
}
