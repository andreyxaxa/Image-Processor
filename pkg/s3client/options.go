package s3client

import "time"

type Option func(c *S3Client)

func ConnAttempts(attempts int) Option {
	return func(c *S3Client) {
		c.connAttempts = attempts
	}
}

func ConnTimeout(timeout time.Duration) Option {
	return func(c *S3Client) {
		c.connTimeout = timeout
	}
}

func Region(region string) Option {
	return func(c *S3Client) {
		c.region = region
	}
}

func UsePathStyle(use bool) Option {
	return func(c *S3Client) {
		c.usePathStyle = use
	}
}
