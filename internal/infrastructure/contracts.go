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
)
