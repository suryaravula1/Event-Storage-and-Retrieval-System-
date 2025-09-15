package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"log-persist-v2/internal/models"
	// "time"

	// "github.com/aws/aws-sdk-go-v2/aws"
	// "github.com/aws/aws-sdk-go-v2/config"
	// "github.com/aws/aws-sdk-go-v2/credentials"
	// "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Uploader encapsulates the S3 client
type S3Uploader struct {
	client *minio.Client
	bucket string
}

// NewS3Uploader initializes an S3Uploader with MinIO or AWS S3 configuration
func NewS3Uploader(endpoint, accessKey, secretKey, bucket string) (*S3Uploader, error) {
	ctx := context.Background()
	log.Printf("initting client")
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),Region: "us-east-1", Secure: false,
	
	})
	log.Printf("innited client")
	if err != nil {
		return nil, fmt.Errorf("failed to load S3 configuration: %w", err)
	}
	log.Printf("making bucket ")
	err = minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: "us-east-1"})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucket)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucket)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", bucket)
	}
	return &S3Uploader{
		client: minioClient,
		bucket: bucket,
	}, nil
}

// UploadToS3 uploads a chunk to the S3-compatible bucket
func (u *S3Uploader) UploadToS3(ctx context.Context, key string, chunk models.Chunk) error {
	select {
	case <-ctx.Done():
		return errors.New("upload timeout")
	default:
		// Upload chunk to S3
		_, err := u.client.PutObject(ctx, 
			 u.bucket,
			key,
			bytes.NewReader([]byte(chunk.Data)),int64(len(chunk.Data)),minio.PutObjectOptions{ContentType: "application/octet-stream"})
		if err != nil {
			log.Printf("Failed to upload chunk %s: %v", chunk.ChunkID, err)
			return err
		}
		log.Printf("Uploaded chunk %s to bucket %s with key %s", chunk.ChunkID, u.bucket, key)
		// time.Sleep(2 * time.Second) // Simulate network delay
		return nil
	}
}
