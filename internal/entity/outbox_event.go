package entity

import (
	"time"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	ID          uuid.UUID  `json:"id"`
	AggregateID uuid.UUID  `json:"aggregate_id"`
	Payload     []byte     `json:"payload"`
	Status      Status     `json:"status"` // pending, processing, processed, failed
	CreatedAt   time.Time  `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	RetryCount  int        `json:"retry_count"`
}
