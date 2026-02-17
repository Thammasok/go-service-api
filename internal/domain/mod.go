package domain

import (
	"dvith.com/go-service-api/internal/domain/common"
	"dvith.com/go-service-api/internal/domain/examples"
	"dvith.com/go-service-api/internal/middleware"
	"github.com/gofiber/fiber/v3"
)

func Init(app *fiber.App) {
	// Group all routes under /api/v1 prefix
	apiV1 := app.Group("/api/v1")

	// Apply centralized error handling middleware to all /api/v1 routes
	apiV1.Use(middleware.ErrorHandler())

	// Register route handlers
	common.Routers(apiV1)

	// Register example handlers (demonstrating error handling)
	examples.RegisterRoutes(apiV1)
}
