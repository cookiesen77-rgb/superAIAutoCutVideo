package storage

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Presigner struct {
	Client *s3.PresignClient
	Bucket string
}

func NewS3Presigner(ctx context.Context, region string, bucket string) (*S3Presigner, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg)
	return &S3Presigner{
		Client: s3.NewPresignClient(client),
		Bucket: bucket,
	}, nil
}

func (s *S3Presigner) PresignPut(ctx context.Context, key string, contentType string, expires time.Duration) (string, error) {
	out, err := s.Client.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expires
	})
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

func (s *S3Presigner) PresignGet(ctx context.Context, key string, expires time.Duration) (string, error) {
	out, err := s.Client.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expires
	})
	if err != nil {
		return "", err
	}
	return out.URL, nil
}
