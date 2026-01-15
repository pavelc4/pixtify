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

func (s *MinIOStorage) InitializeBuckets(ctx context.Context, buckets []string) error {
	for _, bucketName := range buckets {
		err := s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			exists, errBucketExists := s.client.BucketExists(ctx, bucketName)
			if errBucketExists == nil && exists {
				// Bucket already exists, that's fine
			} else {
				return fmt.Errorf("failed to create bucket %s: %w", bucketName, err)
			}
		}

		// Set public policy (readonly)
		policy := fmt.Sprintf(`{"Version": "2012-10-17","Statement": [{"Action": ["s3:GetObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:s3:::%s/*"]}]}`, bucketName)
		err = s.client.SetBucketPolicy(ctx, bucketName, policy)
		if err != nil {
			return fmt.Errorf("failed to set policy for %s: %w", bucketName, err)
		}
	}
	return nil
}

func (s *MinIOStorage) Upload(ctx context.Context, bucket, key string, data io.Reader, size int64, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, bucket, key, data, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload: %w", err)
	}

	if s.cdnURL != "" {
		return fmt.Sprintf("%s/%s", s.cdnURL, key), nil
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
