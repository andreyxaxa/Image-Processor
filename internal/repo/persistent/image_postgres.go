package persistent

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/andreyxaxa/Image-Processor/internal/entity"
	"github.com/andreyxaxa/Image-Processor/pkg/postgres"
	"github.com/andreyxaxa/Image-Processor/pkg/types/errs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	// Table
	imagesTable = "images"

	// Columns
	idColumn           = "id"
	originalKeyColumn  = "original_key"
	processedKeyColumn = "processed_key"
	originalNameColumn = "original_name"
	contentTypeColumn  = "content_type"
	sizeColumn         = "size"
	statusColumn       = "status"
	createdAtColumn    = "created_at"
	processedAtColumn  = "processed_at"
)

type ImageMetadataRepo struct {
	*postgres.Postgres
}

func NewImageMetadataRepo(pg *postgres.Postgres) *ImageMetadataRepo {
	return &ImageMetadataRepo{pg}
}

func (r *ImageMetadataRepo) Create(ctx context.Context, image *entity.Image) error {
	sql, args, err := r.Builder.
		Insert(imagesTable).
		Columns(
			idColumn,
			originalKeyColumn,
			originalNameColumn,
			contentTypeColumn,
			sizeColumn,
			statusColumn,
			createdAtColumn,
		).
		Values(
			image.ID,
			image.OriginalKey,
			image.OriginalName,
			image.ContentType,
			image.Size,
			image.Status,
			image.CreatedAt,
		).ToSql()
	if err != nil {
		return fmt.Errorf("ImageMetadataRepo - Create - r.Builder.ToSql(): %w", err)
	}

	// Pool / Tx
	executor := r.GetExecutor(ctx)

	_, err = executor.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("ImageMetadataRepo - Create - executor.Exec: %w", err)
	}

	return nil
}

func (r *ImageMetadataRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Image, error) {
	sql, args, err := r.Builder.
		Select(
			idColumn,
			originalKeyColumn,
			processedKeyColumn,
			originalNameColumn,
			contentTypeColumn,
			sizeColumn,
			statusColumn,
			createdAtColumn,
			processedAtColumn,
		).
		From(imagesTable).
		Where(squirrel.Eq{idColumn: id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("ImageMetadataRepo - GetByID - r.Builder.ToSql: %w", err)
	}

	executor := r.GetExecutor(ctx)

	var image entity.Image
	err = executor.QueryRow(ctx, sql, args...).Scan(
		&image.ID,
		&image.OriginalKey,
		&image.ProcessedKey,
		&image.OriginalName,
		&image.ContentType,
		&image.Size,
		&image.Status,
		&image.CreatedAt,
		&image.ProcessedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("ImageMetadataRepo - GetByID: %w", errs.ErrRecordNotFound)
		}
		return nil, fmt.Errorf("ImageMetadataRepo - GetByID - executor.QueryRow: %w", err)
	}

	return &image, nil
}

func (r *ImageMetadataRepo) GetProcessedKeyByID(ctx context.Context, id uuid.UUID) (string, string, error) {
	sql, args, err := r.Builder.
		Select(processedKeyColumn, contentTypeColumn).
		From(imagesTable).
		Where(squirrel.And{
			squirrel.Eq{idColumn: id},
			squirrel.Eq{statusColumn: string(entity.Processed)},
		}).
		ToSql()
	if err != nil {
		return "", "", fmt.Errorf("ImageMetadataRepo - GetProcessedKeyByID - r.Builder.ToSql: %w", err)
	}

	executor := r.GetExecutor(ctx)

	var processedKey string
	var contentType string

	err = executor.QueryRow(ctx, sql, args...).Scan(&processedKey, &contentType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", fmt.Errorf("ImageMetadataRepo - GetProcessedKeyByID: %w", errs.ErrRecordNotFound)
		}
		return "", "", fmt.Errorf("ImageMetadataRepo - GetProcessedKeyByID - executor.QueryRow.Scan: %w", err)
	}

	return processedKey, contentType, nil
}

func (r *ImageMetadataRepo) Update(ctx context.Context, image *entity.Image) error {
	sql, args, err := r.Builder.
		Update(imagesTable).
		Set(processedKeyColumn, image.ProcessedKey).
		Set(statusColumn, image.Status).
		Set(processedAtColumn, image.ProcessedAt).
		Where(squirrel.Eq{idColumn: image.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("ImageMetadataRepo - Update - r.Builder.ToSql: %w", err)
	}

	executor := r.GetExecutor(ctx)

	tag, err := executor.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("ImageMetadataRepo - Update - executor.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("ImageMetadataRepo - Update: %w", errs.ErrRecordNotFound)
	}

	return nil
}

func (r *ImageMetadataRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.Builder.
		Delete(imagesTable).
		Where(squirrel.Eq{idColumn: id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("ImageMetadataRepo - Delete - r.Builder.ToSql: %w", err)
	}

	executor := r.GetExecutor(ctx)

	tag, err := executor.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("ImageMetadataRepo - Delete - executor.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("ImageMetadataRepo - Delete: %w", errs.ErrRecordNotFound)
	}

	return nil
}
