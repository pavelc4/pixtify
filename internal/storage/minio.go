package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage struct {
	client *minio.Client
	cdnURL string
}

func NewMinIOStorage(endpoint, accessKey, secretKey, cdnURL string, useSSL bool) (*MinIOStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return &MinIOStorage{client: client, cdnURL: cdnURL}, nil
}

func (s *MinIOStorage) Upload(ctx context.Context, bucket, key string, data io.Reader, size int64, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, bucket, key, data, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload: %w", err)
	}

	if s.cdnURL != "" {
		return fmt.Sprintf("%s/%s/%s", s.cdnURL, bucket, key), nil
	}
	return fmt.Sprintf("http://%s/%s/%s", s.client.EndpointURL().Host, bucket, key), nil
}

func (s *MinIOStorage) Delete(ctx context.Context, bucket, key string) error {
	return s.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
}

func (s *MinIOStorage) GetPresignedURL(ctx context.Context, bucket, key string, expirySeconds int) (string, error) {
	u, err := s.client.PresignedGetObject(ctx, bucket, key, time.Duration(expirySeconds)*time.Second, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
