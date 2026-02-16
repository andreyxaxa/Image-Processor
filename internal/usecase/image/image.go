package image

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/andreyxaxa/Image-Processor/internal/dto"
	"github.com/andreyxaxa/Image-Processor/internal/entity"
	"github.com/andreyxaxa/Image-Processor/internal/repo"
	"github.com/andreyxaxa/Image-Processor/pkg/logger"
	"github.com/google/uuid"
)

type ImageUseCase struct {
	imageRepo          repo.ImageRepo
	metadataRepo       repo.ImageMetadataRepo
	outboxMetadataRepo repo.OutboxImageMetadataRepo
	transactor         repo.Transactor

	logger logger.Interface
}

func New(
	imageRepo repo.ImageRepo,
	metadataRepo repo.ImageMetadataRepo,
	outboxRepo repo.OutboxImageMetadataRepo,
	transactor repo.Transactor,
	l logger.Interface,
) *ImageUseCase {
	return &ImageUseCase{
		imageRepo:          imageRepo,
		metadataRepo:       metadataRepo,
		outboxMetadataRepo: outboxRepo,
		transactor:         transactor,
		logger:             l,
	}
}

func (uc *ImageUseCase) UploadNewImage(
	ctx context.Context,
	data io.Reader,
	originalName string,
	contentType string,
	size int64,
	operation dto.Operation,
) (*entity.Image, error) {
	imageID := uuid.New()
	// TODO: подумать над умным генерированием ключей с датой и форматом файла
	originalKey := fmt.Sprintf("originals/%s", imageID)

	// 1. загружаем в S3
	err := uc.imageRepo.Upload(ctx, originalKey, data, contentType, size)
	if err != nil {
		return nil, fmt.Errorf("ImageUseCase - UploadNewImage - uc.imageRepo.Upload: %w", err)
	}

	image := &entity.Image{
		ID:           imageID,
		OriginalKey:  originalKey,
		OriginalName: originalName,
		ContentType:  contentType,
		Size:         size,
		Status:       entity.Pending,
		CreatedAt:    time.Now(),
	}

	// 2. в единой транзакции
	err = uc.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		// 2.1 записываем метаданные в основную таблицу
		if err := uc.metadataRepo.Create(ctx, image); err != nil {
			return fmt.Errorf("ImageUseCase - UploadNewImage - uc.metadataRepo.Create: %w", err)
		}

		// 2.2 записываем метаданные в аутбокс таблицу
		event, err := uc.createOutboxEvent(imageID, originalKey, contentType, operation)
		if err != nil {
			return fmt.Errorf("ImageUseCase - UploadNewImage - uc.createOutboxEvent: %w", err)
		}
		if err := uc.outboxMetadataRepo.Create(ctx, event); err != nil {
			return fmt.Errorf("ImageUseCase - UploadNewImage - uc.outboxMetadataRepo.Create: %w", err)
		}

		return nil
	})

	// если транзакция не прошла
	if err != nil {
		// удаляем созданный в S3 объект
		deleteErr := uc.imageRepo.Delete(ctx, originalKey)
		if deleteErr != nil {
			uc.logger.Error(deleteErr, "ImageUseCase - UploadNewImage - uc.imageRepo.Delete")
		}
		return nil, fmt.Errorf("ImageUseCase - UploadNewImage - uc.transactor.WithinTransaction: %w", err)
	}

	return image, nil
}

func (uc *ImageUseCase) UploadProcessedImage(ctx context.Context, data []byte, imageID uuid.UUID) error {
	// 1. получим текущие метаданные, чтобы не затереть лишнее
	image, err := uc.metadataRepo.GetByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("ImageUseCase - UploadProcessedImage - uc.metadataRepo.GetByID: %w", err)
	}

	// 2. генерируем ключ и сохраняем в S3
	processedKey := fmt.Sprintf("processed/%s", imageID)
	err = uc.imageRepo.UploadBytes(ctx, processedKey, data, image.ContentType, int64(len(data)))
	if err != nil {
		return fmt.Errorf("ImageUseCase - UploadProcessedImage - uc.imageRepo.UploadBytes: %w", err)
	}

	// 3. модифицируем сущность
	image.ProcessedKey = &processedKey
	image.Status = entity.Processed
	now := time.Now()
	image.ProcessedAt = &now

	// 4. обновляем метаданные
	err = uc.metadataRepo.Update(ctx, image)
	// если не удалось сохранить метаданные
	if err != nil {
		// удалим из S3
		deleteErr := uc.imageRepo.Delete(ctx, processedKey)
		if deleteErr != nil {
			uc.logger.Error(deleteErr, "ImageUseCase - UploadProcessedImage - uc.imageRepo.Delete")
		}
		return fmt.Errorf("ImageUseCase - UploadProcessedImage - uc.metadataRepo.Update: %w", err)
	}

	return nil
}

func (uc *ImageUseCase) DownloadImage(ctx context.Context, key string) (io.ReadCloser, error) {
	body, err := uc.imageRepo.Download(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("ImageUseCase - DownloadImage - uc.imageRepo.Download: %w", err)
	}

	return body, nil
}

func (uc *ImageUseCase) DownloadImageBytes(ctx context.Context, key string) ([]byte, error) {
	b, err := uc.imageRepo.DownloadBytes(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("ImageUseCase - DownloadImageBytes - uc.imageRepo.DownloadBytes: %w", err)
	}

	return b, nil
}

func (uc *ImageUseCase) DeleteImage(ctx context.Context, id uuid.UUID) error {
	// 1. получим ключи для S3
	image, err := uc.metadataRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("ImageUseCase - DeleteImage - uc.metadataRepo.GetByID: %w", err)
	}

	// 2. сначала удалим из основной таблицы в БД (запись в аутбокс удалится каскадно)
	err = uc.metadataRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("ImageUseCase - DeleteImage - uc.metadataRepo.Delete: %w", err)
	}

	// 3. удалим из S3
	// оригинал
	err = uc.imageRepo.Delete(ctx, image.OriginalKey)
	if err != nil {
		uc.logger.Warn("failed to delete key=%s, error=%v", image.OriginalKey, err)
	}

	// обработанное
	if image.ProcessedKey != nil {
		err = uc.imageRepo.Delete(ctx, *image.ProcessedKey)
		if err != nil {
			uc.logger.Warn("failed to delete key=%s, error=%v", image.OriginalKey, err)
		}
	}

	return nil
}

func (uc *ImageUseCase) GetPendingEvents(ctx context.Context, maxRetries, limit int) ([]*entity.OutboxEvent, error) {
	events, err := uc.outboxMetadataRepo.GetPendingEvents(ctx, maxRetries, limit)
	if err != nil {
		return nil, fmt.Errorf("ImageUseCase - GetPendingEvents - uc.outboxMetadataRepo.GetPendingEvents: %w", err)
	}

	return events, nil
}

func (uc *ImageUseCase) MarkAsProcessedBatch(ctx context.Context, events []*entity.OutboxEvent) error {
	var IDs uuid.UUIDs

	for _, event := range events {
		IDs = append(IDs, event.ID)
	}

	err := uc.outboxMetadataRepo.MarkAsProcessedBatch(ctx, IDs)
	if err != nil {
		return fmt.Errorf("ImageUseCase - MarkAsProcessedBatch - uc.outboxMetadataRepo.MarkAsProcessedBatch: %w", err)
	}

	return nil
}

func (uc *ImageUseCase) MarkAsProcessingBatch(ctx context.Context, events []*entity.OutboxEvent) error {
	var IDs uuid.UUIDs

	for _, event := range events {
		IDs = append(IDs, event.ID)
	}

	err := uc.outboxMetadataRepo.MarkAsProcessingBatch(ctx, IDs)
	if err != nil {
		return fmt.Errorf("ImageUseCase - MarkAsProcessingBatch - uc.outboxMetadataRepo.MarkAsProcessingBatch: %w", err)
	}

	return nil
}

func (uc *ImageUseCase) IncrementRetryCountBatch(ctx context.Context, events []*entity.OutboxEvent) error {
	var IDs uuid.UUIDs

	for _, event := range events {
		IDs = append(IDs, event.ID)
	}

	err := uc.outboxMetadataRepo.IncrementRetryCountBatch(ctx, IDs)
	if err != nil {
		return fmt.Errorf("ImageUseCase - IncrementRetryCountBatch - uc.outboxMetadataRepo.IncrementRetryCountBatch: %w", err)
	}

	return nil
}

func (uc *ImageUseCase) GetProcessedKeyByID(ctx context.Context, id uuid.UUID) (string, string, error) {
	pkey, ctype, err := uc.metadataRepo.GetProcessedKeyByID(ctx, id)
	if err != nil {
		return "", "", fmt.Errorf("ImageUseCase - GetProcessedKeyByID - uc.metadataRepo.GetProcessedKeyByID: %w", err)
	}

	return pkey, ctype, nil
}

func (uc *ImageUseCase) MarkMaxRetriesAsFailed(ctx context.Context, maxRetries int) error {
	err := uc.outboxMetadataRepo.MarkMaxRetriesAsFailed(ctx, maxRetries)
	if err != nil {
		return fmt.Errorf("ImageUseCase - MarkMaxRetriesAsFailed - uc.outboxMetadataRepo.MarkMaxRetriesAsFailed: %w", err)
	}

	return nil
}

func (uc *ImageUseCase) CleanupOutbox(ctx context.Context) error {
	count, err := uc.outboxMetadataRepo.DeleteOldProcessedAndFailed(ctx)
	if err != nil {
		return fmt.Errorf("ImageUseCase - CleanupOutbox - uc.CleanupOutbox.DeleteOldProcessedAndFailed: %w", err)
	}

	if count > 0 {
		uc.logger.Info("deleted old events, count = %d", count)
	}

	return nil
}
