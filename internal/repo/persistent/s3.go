package persistent

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/andreyxaxa/Image-Processor/pkg/s3client"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ImageRepo struct {
	*s3client.S3Client
	bucket string
}

func NewImageRepo(s3c *s3client.S3Client, bucket string) *ImageRepo {
	return &ImageRepo{s3c, bucket}
}

func (r *ImageRepo) Upload(ctx context.Context, key string, data io.Reader, contentType string, size int64) error {
	_, err := r.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(r.bucket),
		Key:           aws.String(key),
		Body:          data,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})
	if err != nil {
		return fmt.Errorf("ImageRepo - Upload - r.c.Client.PutObject: %w", err)
	}

	return nil
}

func (r *ImageRepo) UploadBytes(ctx context.Context, key string, data []byte, contentType string, size int64) error {
	b := bytes.NewReader(data)

	_, err := r.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(r.bucket),
		Key:           aws.String(key),
		Body:          b,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})
	if err != nil {
		return fmt.Errorf("ImageRepo - Upload - r.c.Client.PutObject: %w", err)
	}

	return nil
}

func (r *ImageRepo) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := r.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("ImageRepo - Download - r.c.Client.GetObject: %w", err)
	}

	return result.Body, nil
}

func (r *ImageRepo) DownloadBytes(ctx context.Context, key string) ([]byte, error) {
	result, err := r.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("ImageRepo - Download - r.c.Client.GetObject: %w", err)
	}
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("ImageRepo - Download - io.ReadAll: %w", err)
	}

	return b, nil
}

func (r *ImageRepo) Delete(ctx context.Context, key string) error {
	_, err := r.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("ImageRepo - Delete - r.c.Client.DeleteObject: %w", err)
	}

	return nil
}
