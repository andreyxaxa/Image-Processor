package s3client

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
	_defaultRegion       = "garage"
)

type S3Client struct {
	connAttempts int
	connTimeout  time.Duration

	endpoint     string
	region       string
	accessKey    string
	secretKey    string
	usePathStyle bool

	Client *s3.Client
}

func New(ctx context.Context, endpoint, accessKey, secretKey string, opts ...Option) (*S3Client, error) {
	s3c := &S3Client{
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
		region:       _defaultRegion,
		endpoint:     endpoint,
		accessKey:    accessKey,
		secretKey:    secretKey,
		usePathStyle: true,
	}

	for _, opt := range opts {
		opt(s3c)
	}

	var err error
	for s3c.connAttempts > 0 {
		err = s3c.connect(ctx)
		if err == nil {
			break
		}

		log.Printf("S3 is trying to connect, attempts left: %d", s3c.connAttempts)

		time.Sleep(s3c.connTimeout)

		s3c.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("S3Client - New - connAttempts == 0: %w", err)
	}

	return s3c, nil
}

func (s *S3Client) connect(ctx context.Context) error {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(s.region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(s.accessKey, s.secretKey, ""),
		),
	)
	if err != nil {
		return fmt.Errorf("S3Client - config.LoadDefaultConfig: %w", err)
	}

	s.Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = s.usePathStyle
		o.BaseEndpoint = aws.String(s.endpoint)
	})

	// check connection
	_, err = s.Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("S3Client - s.Client.ListBuckets: %w", err)
	}

	return nil
}
