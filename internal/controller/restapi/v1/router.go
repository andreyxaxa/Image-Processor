package v1

import (
	"github.com/andreyxaxa/Image-Processor/internal/usecase"
	"github.com/andreyxaxa/Image-Processor/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

func NewImageRoutes(apiV1Group fiber.Router, img usecase.ImageUseCase, l logger.Interface) {
	r := &V1{img: img, logger: l}

	{
		// API
		apiV1Group.Post("/upload", r.processImage)
		apiV1Group.Get("/image/:id", r.getProcessedImage)
		apiV1Group.Delete("/image/:id", r.deleteImage)

		// UI
		apiV1Group.Get("/", r.showUI)
	}
}
