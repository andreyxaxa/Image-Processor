package v1

import (
	"github.com/andreyxaxa/Image-Processor/internal/usecase"
	"github.com/andreyxaxa/Image-Processor/pkg/logger"
)

type V1 struct {
	img    usecase.ImageUseCase
	logger logger.Interface
}
