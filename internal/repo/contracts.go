package repo

import (
	"context"
	"io"

	"github.com/andreyxaxa/Image-Processor/internal/entity"
	"github.com/google/uuid"
)

type (
	ImageRepo interface {
		Upload(ctx context.Context, key string, data io.Reader, contentType string, size int64) error
		UploadBytes(ctx context.Context, key string, data []byte, contentType string, size int64) error
		Download(ctx context.Context, key string) (io.ReadCloser, error)
		DownloadBytes(ctx context.Context, key string) ([]byte, error)
		Delete(ctx context.Context, key string) error
	}

	ImageMetadataRepo interface {
		Create(ctx context.Context, image *entity.Image) error
		GetByID(ctx context.Context, id uuid.UUID) (*entity.Image, error)
		GetProcessedKeyByID(ctx context.Context, id uuid.UUID) (string, string, error)
		Update(ctx context.Context, image *entity.Image) error
		Delete(ctx context.Context, id uuid.UUID) error
	}

	OutboxImageMetadataRepo interface {
		Create(ctx context.Context, event *entity.OutboxEvent) error
		GetPendingEvents(ctx context.Context, maxRetries int, limit int) ([]*entity.OutboxEvent, error)
		MarkAsProcessingBatch(ctx context.Context, IDs uuid.UUIDs) error
		MarkAsProcessedBatch(ctx context.Context, IDs uuid.UUIDs) error
		MarkAsFailedBatch(ctx context.Context, IDs uuid.UUIDs) error
		IncrementRetryCountBatch(ctx context.Context, IDs uuid.UUIDs) error
		MarkMaxRetriesAsFailed(ctx context.Context, maxRetries int) error
		DeleteOldProcessedAndFailed(ctx context.Context) (int64, error)
	}

	Transactor interface {
		WithinTransaction(ctx context.Context, f func(ctx context.Context) error) error
	}
)
