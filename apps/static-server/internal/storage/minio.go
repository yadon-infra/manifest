package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

type MinioStorage struct {
	client *minio.Client
	logger *logrus.Logger
}

func NewMinioStorage(endpoint, accessKey, secretKey string, useSSL bool, logger *logrus.Logger) (*MinioStorage, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MinIO client: %w", err)
	}

	return &MinioStorage{
		client: minioClient,
		logger: logger,
	}, nil
}

func (m *MinioStorage) GetObject(ctx context.Context, bucket, objectName string) (io.ReadCloser, error) {
	object, err := m.client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return object, nil
}

func (m *MinioStorage) ObjectExists(ctx context.Context, bucket, objectName string) bool {
	_, err := m.client.StatObject(ctx, bucket, objectName, minio.StatObjectOptions{})
	return err == nil
}