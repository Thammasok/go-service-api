package domain

import (
	"dvith.com/go-service-api/internal/config"
	"dvith.com/go-service-api/internal/domain/authentication"
	"dvith.com/go-service-api/internal/domain/common"
	"dvith.com/go-service-api/internal/domain/examples"
	"dvith.com/go-service-api/internal/middleware"
	"dvith.com/go-service-api/pkg/database"
	"github.com/gofiber/fiber/v3"
)

func Init(app *fiber.App, db *database.DBPool, cfg config.Config) {
	// Group all routes under /api/v1 prefix
	apiV1 := app.Group("/api/v1")

	// Apply centralized error handling middleware to all /api/v1 routes
	apiV1.Use(middleware.ErrorHandler())

	// Register route handlers
	common.Routers(apiV1)
	authentication.Routers(apiV1, db, cfg)

	// Register example handlers (demonstrating error handling)
	examples.RegisterRoutes(apiV1)
}
