package image

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/andreyxaxa/Image-Processor/internal/dto"
	"github.com/andreyxaxa/Image-Processor/internal/entity"
	"github.com/google/uuid"
)

func (uc *ImageUseCase) createOutboxEvent(
	imageID uuid.UUID,
	originalKey string,
	contentType string,
	operation dto.Operation,
) (*entity.OutboxEvent, error) {
	payload := map[string]interface{}{
		"id":           imageID,
		"original_key": originalKey,
		"content_type": contentType,
		"operation":    operation.Operation,
		"width":        operation.Width,
		"height":       operation.Height,
		"text":         operation.Text,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("ImageUseCase - createOutboxEvent - json.Marshal: %w", err)
	}

	return &entity.OutboxEvent{
		ID:          uuid.New(),
		AggregateID: imageID,
		Payload:     b,
		Status:      entity.Pending,
		CreatedAt:   time.Now(),
		RetryCount:  0,
	}, nil
}
