package domain

import (
	"dvith.com/go-service-api/internal/domain/common"
	"github.com/gofiber/fiber/v3"
)

func Init(app *fiber.App) {
	// Group all routes under /api/v1 prefix
	apiV1 := app.Group("/api/v1")
	common.Routers(apiV1)
}
