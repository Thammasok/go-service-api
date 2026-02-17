package examples

import "github.com/gofiber/fiber/v3"

// RegisterRoutes registers all example handler routes.
func RegisterRoutes(router fiber.Router) {
	examples := router.Group("/examples")

	// User management examples
	examples.Post("/users", ExampleCreateUserHandler)
	examples.Get("/users/:id", ExampleGetUserHandler)

	// Authentication example
	examples.Get("/protected", ExampleAuthenticatedHandler)

	// Database error example
	examples.Get("/data", ExampleDatabaseErrorHandler)

	// Panic recovery example
	examples.Get("/panic", ExamplePanicHandler)
}
