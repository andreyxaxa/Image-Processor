package usecase

import (
	"context"
	"io"

	"github.com/andreyxaxa/Image-Processor/internal/dto"
	"github.com/andreyxaxa/Image-Processor/internal/entity"
	"github.com/google/uuid"
)

type (
	ImageUseCase interface {
		UploadNewImage(
			ctx context.Context,
			data io.Reader,
			originalName string,
			contentType string,
			size int64,
			operation dto.Operation,
		) (*entity.Image, error)
		UploadProcessedImage(ctx context.Context, data []byte, imageID uuid.UUID) error
		DownloadImage(ctx context.Context, key string) (io.ReadCloser, error)
		DownloadImageBytes(ctx context.Context, key string) ([]byte, error)
		DeleteImage(ctx context.Context, id uuid.UUID) error
		GetProcessedKeyByID(ctx context.Context, id uuid.UUID) (string, string, error)
		GetPendingEvents(ctx context.Context, maxRetries, limit int) ([]*entity.OutboxEvent, error)
		MarkAsProcessingBatch(ctx context.Context, events []*entity.OutboxEvent) error
		MarkAsProcessedBatch(ctx context.Context, events []*entity.OutboxEvent) error
		IncrementRetryCountBatch(ctx context.Context, events []*entity.OutboxEvent) error
		MarkMaxRetriesAsFailed(ctx context.Context, maxRetries int) error
		CleanupOutbox(ctx context.Context) error
	}

	ImageProcessorUseCase interface {
		Process(ctx context.Context, contentType string, task dto.Task) ([]byte, error)
	}
)
