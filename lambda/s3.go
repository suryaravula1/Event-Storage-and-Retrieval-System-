package main
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Downloader struct {
	Client *s3.Client
	Bucket string
}

// NewS3Uploader initializes an S3Uploader with MinIO or AWS S3 configuration
func NewS3Downloader(endpoint, accessKey, secretKey, bucket string) (*S3Downloader, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID {
			return aws.Endpoint{
				URL:           endpoint,
				SigningRegion: "us-east-1",
			}, nil
		}
		return aws.Endpoint{}, fmt.Errorf("unknown service: %s", service)
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load S3 configuration: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3Downloader{
		Client: client,
		Bucket: bucket,
	}, nil
}


func  Download(){}

func (downloader *S3Downloader) DownloadS3File(bucket, key string) ([]byte, error) {
	

	req, _ := downloader.Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	buffer := bytes.NewBuffer([]byte{})
	err := req.Send()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
