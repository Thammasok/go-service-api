package common

import (
	"dvith.com/go-service-api/internal/domain/common/health"
	"dvith.com/go-service-api/internal/domain/common/home"
	"github.com/gofiber/fiber/v3"
)

func Routers(app fiber.Router) {
	app.Get("/", home.HomeHandler)
	app.Get("/health", health.HealthHandler)
}
