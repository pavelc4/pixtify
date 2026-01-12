package main

import (
	"context"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	endpoint := "localhost:9000"
	accessKeyID := "pixtify_admin"
	secretAccessKey := "pixtify_secret_key"
	useSSL := false

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	buckets := []string{"pixtify-originals", "pixtify-thumbnails"}

	for _, bucketName := range buckets {
		ctx := context.Background()
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			// Check to see if we already own this bucket (which happens if it exists)
			exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
			if errBucketExists == nil && exists {
				log.Printf("We already own %s\n", bucketName)
			} else {
				log.Fatalln(err)
			}
		} else {
			log.Printf("Successfully created %s\n", bucketName)
		}

		// Set public policy for thumbnails (readonly)
		if bucketName == "pixtify-thumbnails" {
			policy := `{"Version": "2012-10-17","Statement": [{"Action": ["s3:GetObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:s3:::pixtify-thumbnails/*"]}]}`
			err = minioClient.SetBucketPolicy(ctx, bucketName, policy)
			if err != nil {
				log.Printf("Failed to set policy for %s: %v\n", bucketName, err)
			} else {
				log.Printf("Set public policy for %s\n", bucketName)
			}
		}
	}
}
