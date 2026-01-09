package storage

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Uploader struct {
	Client *manager.Uploader
	Bucket string
	Region string
}

func NewS3Uploader(ctx context.Context, region string, bucket string) (*S3Uploader, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg)
	return &S3Uploader{
		Client: manager.NewUploader(client),
		Bucket: bucket,
		Region: region,
	}, nil
}

func (u *S3Uploader) Upload(ctx context.Context, key string, contentType string, body io.Reader) (string, error) {
	_, err := u.Client.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(u.Bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		Body:        body,
	})
	if err != nil {
		return "", err
	}
	return "https://" + u.Bucket + ".s3." + u.Region + ".amazonaws.com/" + key, nil
}
