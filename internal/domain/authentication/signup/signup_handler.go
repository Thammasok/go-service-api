package signup

import (
	"dvith.com/go-service-api/pkg/database"
	"github.com/gofiber/fiber/v3"
)

// SignupHandler handles user signup requests
func SignupHandler(db *database.DBPool) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Parse signup request
		var req SignupRequest
		if err := c.Bind().Body(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate request fields
		validationErrors := ValidateSignupRequest(&req)
		if len(validationErrors) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":  "Validation failed",
				"errors": validationErrors,
			})
		}

		// Create repository and service
		repo := NewSignupRepository(db)
		service := NewSignupService(repo)

		// Register user (hash password and save to database)
		user, err := service.RegisterUser(c.Context(), &req)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Return success response with user data
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "User registered successfully",
			"user": fiber.Map{
				"id":        user.ID,
				"email":     user.Email,
				"fullName":  user.FullName,
				"username":  user.Username,
				"isActive":  user.IsActive,
				"createdAt": user.CreatedAt,
			},
		})
	}
}
