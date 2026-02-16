package kafka

import "github.com/google/uuid"

type ImageEventPayload struct {
	ID          uuid.UUID `json:"id"`
	OriginalKey string    `json:"original_key"`
	ContentType string    `json:"content_type"`
	Operation   string    `json:"operation"`
	Width       *int      `json:"width,omitempty"`
	Height      *int      `json:"height,omitempty"`
}
