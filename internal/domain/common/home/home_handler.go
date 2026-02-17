package home

import (
	"github.com/gofiber/fiber/v3"
)

func HomeHandler(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Welcome to the Go Service API!"})
}
