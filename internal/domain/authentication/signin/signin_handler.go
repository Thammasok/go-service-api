package signin

import (
	"dvith.com/go-service-api/internal/config"
	"dvith.com/go-service-api/internal/security/token"
	"dvith.com/go-service-api/pkg/database"
	"github.com/gofiber/fiber/v3"
)

// SigninHandler handles user signin requests
func SigninHandler(db *database.DBPool, cfg config.Config) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Parse signin request
		var req SigninRequest
		if err := c.Bind().Body(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate request fields
		validationErrors := ValidateSigninRequest(&req)
		if len(validationErrors) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":  "Validation failed",
				"errors": validationErrors,
			})
		}

		// Create repository and service with token manager
		repo := NewSigninRepository(db)
		tokenManager := token.NewTokenManager(token.TokenConfig{
			SecretKey:       cfg.JWTSecretKey,
			ExpirationTime:  cfg.JWTExpirationTime,
			RefreshDuration: cfg.JWTRefreshDuration,
			Issuer:          cfg.JWTIssuer,
		})
		service := NewSigninService(repo, tokenManager)

		// Login user and generate tokens
		response, err := service.LoginUser(c.Context(), &req)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Return success response with user data and tokens
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "User logged in successfully",
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
