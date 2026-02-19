package refreshtoken

import (
	"dvith.com/go-service-api/internal/config"
	"dvith.com/go-service-api/internal/middleware"
	"dvith.com/go-service-api/internal/security/token"
	"dvith.com/go-service-api/pkg/database"
	"dvith.com/go-service-api/pkg/logger"
	"github.com/gofiber/fiber/v3"
)

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenHandler handles refresh token requests
func RefreshTokenHandler(db *database.DBPool, cfg config.Config) fiber.Handler {
	tm := token.NewTokenManager(token.TokenConfig{
		SecretKey:       cfg.JWTSecretKey,
		ExpirationTime:  cfg.JWTExpirationTime,
		RefreshDuration: cfg.JWTRefreshDuration,
		Issuer:          cfg.JWTIssuer,
	})

	return func(c fiber.Ctx) error {
		var req RefreshTokenRequest

		// Parse request body
		if err := c.Bind().Body(&req); err != nil {
			logger.Warn("invalid refresh token request", map[string]any{
				"error": err.Error(),
			})
			return middleware.ValidationErrorResponse(c, "invalid request body")
		}

		// Validate refresh token
		claims, err := tm.ValidateRefreshToken(req.RefreshToken)
		if err != nil {
			logger.Warn("invalid or expired refresh token", map[string]any{
				"error": err.Error(),
			})
			return middleware.AuthErrorResponse(c, "invalid or expired refresh token")
		}

		// Generate new access token
		newAccessToken, err := tm.GenerateAccessToken(claims.UserID)
		if err != nil {
			logger.Error("failed to generate access token", map[string]any{
				"user_id": claims.UserID.String(),
				"error":   err.Error(),
			})
			return middleware.InternalErrorResponse(c, "failed to generate access token")
		}

		logger.Info("refresh token used", map[string]any{
			"user_id": claims.UserID.String(),
		})

		return c.Status(fiber.StatusOK).JSON(token.TokenPair{
			AccessToken:  newAccessToken,
			RefreshToken: req.RefreshToken, // Return same refresh token
			TokenType:    "Bearer",
			ExpiresIn:    int64(cfg.JWTExpirationTime.Seconds()),
		})
	}
}
