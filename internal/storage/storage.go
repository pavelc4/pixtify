package storage

import (
	"context"
	"io"
)

type Service interface {
	Upload(ctx context.Context, bucket, key string, data io.Reader, size int64, contentType string) (string, error)
	Delete(ctx context.Context, bucket, key string) error
	GetPresignedURL(ctx context.Context, bucket, key string, expirySeconds int) (string, error)
}
