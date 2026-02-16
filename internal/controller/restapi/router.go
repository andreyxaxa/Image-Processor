package restapi

import (
	"github.com/andreyxaxa/Image-Processor/config"
	v1 "github.com/andreyxaxa/Image-Processor/internal/controller/restapi/v1"
	"github.com/andreyxaxa/Image-Processor/internal/usecase"
	"github.com/andreyxaxa/Image-Processor/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// @title Image processor
// @version 1.0.0
// @host localhost:8080
// @BasePath /v1
func NewRouter(app *fiber.App, cfg *config.Config, img usecase.ImageUseCase, l logger.Interface) {
	// Swagger
	if cfg.Swagger.Enabled {
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// Routers
	apiV1Group := app.Group("/v1")
	{
		v1.NewImageRoutes(apiV1Group, img, l)
	}
}
