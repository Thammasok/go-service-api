package signup

import (
	"dvith.com/go-service-api/internal/config"
	"dvith.com/go-service-api/internal/security/token"
	"dvith.com/go-service-api/pkg/database"
	"github.com/gofiber/fiber/v3"
)

// SignupHandler handles user signup requests
func SignupHandler(db *database.DBPool, cfg config.Config) fiber.Handler {
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

		// Create repository and service with token manager
		repo := NewSignupRepository(db)
		tokenManager := token.NewTokenManager(token.TokenConfig{
			SecretKey:       cfg.JWTSecretKey,
			ExpirationTime:  cfg.JWTExpirationTime,
			RefreshDuration: cfg.JWTRefreshDuration,
			Issuer:          cfg.JWTIssuer,
		})
		service := NewSignupService(repo, tokenManager)

		// Register user (hash password and save to database)
		response, err := service.RegisterUser(c.Context(), &req)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Return success response with user data and tokens
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "User registered successfully",
			"user": fiber.Map{
				"id":        response.User.ID,
				"email":     response.User.Email,
				"fullName":  response.User.FullName,
				"username":  response.User.Username,
				"isActive":  response.User.IsActive,
				"createdAt": response.User.CreatedAt,
			},
			"access_token":  response.AccessToken,
			"refresh_token": response.RefreshToken,
			"token_type":    response.TokenType,
			"expires_in":    response.ExpiresIn,
		})
	}
}
