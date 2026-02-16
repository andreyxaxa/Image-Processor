package infrastructure

import (
	"context"

	"github.com/andreyxaxa/Image-Processor/internal/entity"
)

type (
	EventsSender interface {
		SendEvents(ctx context.Context, events []*entity.OutboxEvent) error
		Close() error
	}

	ImageProcessor interface {
		Resize(ctx context.Context, contentType string, data []byte, width, height int) ([]byte, error)
		Thumbnail(ctx context.Context, contentType string, data []byte) ([]byte, error)
		Watermark(ctx context.Context, contentType string, data []byte) ([]byte, error)
	}
)
