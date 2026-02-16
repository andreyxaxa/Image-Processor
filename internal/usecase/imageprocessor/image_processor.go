package imageprocessor

import (
	"context"
	"fmt"

	"github.com/andreyxaxa/Image-Processor/internal/dto"
	"github.com/andreyxaxa/Image-Processor/internal/infrastructure"
	"github.com/andreyxaxa/Image-Processor/pkg/types/errs"
)

const (
	resize    = "resize"
	watermark = "watermark"
	thumbnail = "thumbnail"
)

type ImageProcessorUseCase struct {
	p infrastructure.ImageProcessor
}

func New(p infrastructure.ImageProcessor) *ImageProcessorUseCase {
	return &ImageProcessorUseCase{p}
}

func (uc *ImageProcessorUseCase) Process(ctx context.Context, contentType string, task dto.Task) ([]byte, error) {
	var result []byte
	var err error

	switch task.Operation {
	case resize:
		result, err = uc.p.Resize(ctx, contentType, task.Data, *task.Width, *task.Height)
	case watermark:
		result, err = uc.p.Watermark(ctx, contentType, task.Data, *task.Text)
	case thumbnail:
		result, err = uc.p.Thumbnail(ctx, contentType, task.Data)
	default:
		return nil, fmt.Errorf("ImageProcessorUseCase - Process: %w", errs.ErrUnknownOperation)
	}

	if err != nil {
		return nil, fmt.Errorf("ImageProcessorUseCase - Process: %w", err)
	}

	return result, nil
}
