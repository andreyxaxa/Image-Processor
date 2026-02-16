package entity

import (
	"time"

	"github.com/google/uuid"
)

type Image struct {
	ID uuid.UUID `json:"id"`

	OriginalKey  string  `json:"original_key"`
	ProcessedKey *string `json:"processed_key,omitempty"`

	OriginalName string `json:"original_name"`
	ContentType  string `json:"content_type"`
	Size         int64  `json:"size"`
	Status       Status `json:"status"` // pending, processed

	CreatedAt   time.Time  `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}
