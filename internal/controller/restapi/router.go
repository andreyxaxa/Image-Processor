package restapi

import (
	v1 "github.com/andreyxaxa/Image-Processor/internal/controller/restapi/v1"
	"github.com/andreyxaxa/Image-Processor/internal/usecase"
	"github.com/andreyxaxa/Image-Processor/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

func NewRouter(app *fiber.App, img usecase.ImageUseCase, l logger.Interface) {
	// Routers
	apiV1Group := app.Group("/v1")
	{
		v1.NewImageRoutes(apiV1Group, img, l)
	}
}
