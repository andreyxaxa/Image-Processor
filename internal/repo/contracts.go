package repo

import (
	"context"
	"io"
)

type (
	ImageRepo interface {
		Upload(ctx context.Context, key string, data io.Reader, contentType string, size int64) error
		UploadBytes(ctx context.Context, key string, data []byte, contentType string, size int64) error
		Download(ctx context.Context, key string) (io.ReadCloser, error)
		DownloadBytes(ctx context.Context, key string) ([]byte, error)
		Delete(ctx context.Context, key string) error
	}
)
