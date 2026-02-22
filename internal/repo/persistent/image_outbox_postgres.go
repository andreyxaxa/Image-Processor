package persistent

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/andreyxaxa/Image-Processor/internal/entity"
	"github.com/andreyxaxa/Image-Processor/pkg/postgres"
	"github.com/andreyxaxa/Image-Processor/pkg/types/errs"
	"github.com/google/uuid"
)

const (
	// Table
	outboxTable = "images_outbox"

	// Columns
	outboxIDColumn          = "id"
	outboxAggregateIDColumn = "aggregate_id"
	outboxPayloadColumn     = "payload"
	outboxStatusColumn      = "status"
	outboxCreatedAtColumn   = "created_at"
	outboxProcessedAtColumn = "processed_at"
	outboxRetryCountColumn  = "retry_count"
)

type OutboxImageMetadataRepo struct {
	*postgres.Postgres
}

func NewOutboxImageMetadataRepo(pg *postgres.Postgres) *OutboxImageMetadataRepo {
	return &OutboxImageMetadataRepo{pg}
}

func (r *OutboxImageMetadataRepo) Create(ctx context.Context, event *entity.OutboxEvent) error {
	sql, args, err := r.Builder.
		Insert(outboxTable).
		Columns(
			outboxIDColumn,
			outboxAggregateIDColumn,
			outboxPayloadColumn,
			outboxStatusColumn,
			outboxCreatedAtColumn,
			outboxRetryCountColumn,
		).
		Values(
			event.ID,
			event.AggregateID,
			event.Payload,
			event.Status,
			event.CreatedAt,
			event.RetryCount,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("OutboxImageMetadataRepo - Create - r.Builder.ToSql: %w", err)
	}

	executor := r.GetExecutor(ctx)

	_, err = executor.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("OutboxImageMetadataRepo - Create - executor.Exec: %w", err)
	}

	return nil
}

func (r *OutboxImageMetadataRepo) GetPendingEvents(ctx context.Context, limit int, maxRetries int) ([]*entity.OutboxEvent, error) {
	sql, args, err := r.Builder.
		Select(
			outboxIDColumn,
			outboxAggregateIDColumn,
			outboxPayloadColumn,
			outboxStatusColumn,
			outboxCreatedAtColumn,
			outboxProcessedAtColumn,
			outboxRetryCountColumn,
		).
		From(outboxTable).
		Where(squirrel.And{
			squirrel.Eq{outboxStatusColumn: entity.Pending},
			squirrel.Lt{outboxRetryCountColumn: maxRetries},
		}).
		OrderBy("created_at ASC").
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("OutboxImageMetadataRepo - GetPendingEvents - r.Builder.ToSql: %w", err)
	}

	executor := r.GetExecutor(ctx)

	rows, err := executor.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("OutboxImageMetadataRepo - GetPendingEvents - executor.Query: %w", err)
	}
	defer rows.Close()

	events := make([]*entity.OutboxEvent, 0, limit)
	for rows.Next() {
		var event entity.OutboxEvent
		err = rows.Scan(
			&event.ID,
			&event.AggregateID,
			&event.Payload,
			&event.Status,
			&event.CreatedAt,
			&event.ProcessedAt,
			&event.RetryCount,
		)
		if err != nil {
			return nil, fmt.Errorf("OutboxImageMetadataRepo - GetPendingEvents - rows.Scan: %w", err)
		}
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("OutboxImageMetadataRepo - GetPendingEvents - rows.Err: %w", err)
	}

	return events, nil
}

func (r *OutboxImageMetadataRepo) MarkAsProcessingBatch(ctx context.Context, IDs uuid.UUIDs) error {
	now := time.Now()

	sql, args, err := r.Builder.
		Update(outboxTable).
		Set(outboxStatusColumn, entity.Processing).
		Set(outboxProcessedAtColumn, now).
		Where(squirrel.Eq{outboxIDColumn: IDs}).
		ToSql()
	if err != nil {
		return fmt.Errorf("OutboxImageMetadataRepo - MarkAsProcessingBatch - r.Builder.ToSql: %w", err)
	}

	executor := r.GetExecutor(ctx)

	tag, err := executor.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("OutboxImageMetadataRepo - MarkAsProcessingBatch - executor.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("OutboxImageMetadataRepo - MarkAsProcessingBatch: %w", errs.ErrRecordNotFound)
	}

	return nil
}

func (r *OutboxImageMetadataRepo) MarkAsProcessedBatch(ctx context.Context, IDs uuid.UUIDs) error {
	now := time.Now()

	sql, args, err := r.Builder.
		Update(outboxTable).
		Set(outboxStatusColumn, entity.Processed).
		Set(outboxProcessedAtColumn, now).
		Where(squirrel.Eq{outboxIDColumn: IDs}).
		ToSql()
	if err != nil {
		return fmt.Errorf("OutboxImageMetadataRepo - MarkAsProcessedBatch - r.Builder.ToSql: %w", err)
	}

	executor := r.GetExecutor(ctx)

	tag, err := executor.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("OutboxImageMetadataRepo - MarkAsProcessedBatch - executor.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("OutboxImageMetadataRepo - MarkAsProcessedBatch: %w", errs.ErrRecordNotFound)
	}

	return nil
}

func (r *OutboxImageMetadataRepo) MarkAsFailedBatch(ctx context.Context, IDs uuid.UUIDs) error {
	sql, args, err := r.Builder.
		Update(outboxTable).
		Set(outboxStatusColumn, entity.Failed).
		Where(squirrel.Eq{outboxIDColumn: IDs}).
		ToSql()
	if err != nil {
		return fmt.Errorf("OutboxImageMetadataRepo - MarkAsFailedBatch - r.Builder.ToSql: %w", err)
	}

	executor := r.GetExecutor(ctx)

	tag, err := executor.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("OutboxImageMetadataRepo - MarkAsFailedBatch - executor.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("OutboxImageMetadataRepo - MarkAsFailedBatch: %w", errs.ErrRecordNotFound)
	}

	return nil
}

func (r *OutboxImageMetadataRepo) MarkMaxRetriesAsFailed(ctx context.Context, maxRetries int) error {
	sql, args, err := r.Builder.
		Update(outboxTable).
		Set(outboxStatusColumn, entity.Failed).
		Where(squirrel.And{
			squirrel.Eq{outboxStatusColumn: string(entity.Pending)},
			squirrel.GtOrEq{outboxRetryCountColumn: maxRetries},
		}).
		ToSql()
	if err != nil {
		return fmt.Errorf("OutboxImageMetadataRepo - MarkMaxRetriesAsFailed - r.Builder.ToSql: %w", err)
	}

	executor := r.GetExecutor(ctx)

	_, err = executor.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("OutboxImageMetadataRepo - MarkMaxRetriesAsFailed - executor.Exec: %w", err)
	}

	return nil
}

func (r *OutboxImageMetadataRepo) IncrementRetryCountBatch(ctx context.Context, IDs uuid.UUIDs) error {
	sql, args, err := r.Builder.
		Update(outboxTable).
		Set(outboxRetryCountColumn, squirrel.Expr(outboxRetryCountColumn+" + 1")).
		Set(outboxStatusColumn, entity.Pending).
		Where(squirrel.Eq{outboxIDColumn: IDs}).
		ToSql()
	if err != nil {
		return fmt.Errorf("OutboxImageMetadataRepo - IncrementRetryCountBatch - r.Builder.ToSql: %w", err)
	}

	executor := r.GetExecutor(ctx)

	tag, err := executor.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("OutboxImageMetadataRepo - IncrementRetryCountBatch - executor.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("OutboxImageMetadataRepo - IncrementRetryCountBatch: %w", errs.ErrRecordNotFound)
	}

	return nil
}

func (r *OutboxImageMetadataRepo) DeleteOldProcessedAndFailed(ctx context.Context) (int64, error) {
	sql, args, err := r.Builder.
		Delete(outboxTable).
		Where(squirrel.Or{
			squirrel.Eq{outboxStatusColumn: string(entity.Processed)},
			squirrel.Eq{outboxStatusColumn: string(entity.Processed)},
		}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("OutboxImageMetadataRepo - DeleteOldProcessedAndFailed - r.Builder.ToSql: %w", err)
	}

	executor := r.GetExecutor(ctx)
	tag, err := executor.Exec(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("OutboxImageMetadataRepo - DeleteOldProcessedAndFailed - executor.Exec: %w", err)
	}

	return tag.RowsAffected(), nil
}
