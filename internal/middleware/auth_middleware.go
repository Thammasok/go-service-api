package middleware

import (
	"fmt"
	"strings"

	"dvith.com/go-service-api/internal/security/token"
	"dvith.com/go-service-api/pkg/logger"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// Context key constants
const (
	ContextKeyUserID = "user_id"
)

// AuthMiddleware validates JWT access token from Authorization header
func AuthMiddleware(tm *token.TokenManager) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Extract bearer token from authorization header
		authHeader := c.Get("Authorization", "")
		if authHeader == "" {
			logger.Warn("missing authorization header", map[string]any{
				"path":   c.Path(),
				"method": c.Method(),
			})
			return AuthErrorResponse(c, "missing authorization header")
		}

		// Extract token from "Bearer <token>" format
		tokenString, err := extractBearerToken(authHeader)
		if err != nil {
			logger.Warn("invalid authorization header format", map[string]any{
				"path":  c.Path(),
				"error": err.Error(),
			})
			return AuthErrorResponse(c, "invalid authorization header format")
		}

		// Validate access token
		claims, err := tm.ValidateAccessToken(tokenString)
		if err != nil {
			logger.Warn("invalid or expired access token", map[string]any{
				"path":  c.Path(),
				"error": err.Error(),
			})
			return AuthErrorResponse(c, "invalid or expired access token")
		}

		// Store user ID in context for use in handlers
		c.Locals(ContextKeyUserID, claims.UserID)

		logger.Debug("user authenticated", map[string]any{
			"user_id": claims.UserID.String(),
			"path":    c.Path(),
		})

		return c.Next()
	}
}

// extractBearerToken extracts the token from "Bearer <token>" header
func extractBearerToken(authHeader string) (string, error) {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid bearer token format")
	}

	if parts[1] == "" {
		return "", fmt.Errorf("empty token")
	}

	return parts[1], nil
}

// GetUserIDFromContext retrieves the user ID stored in context by AuthMiddleware
func GetUserIDFromContext(c fiber.Ctx) (uuid.UUID, error) {
	val := c.Locals(ContextKeyUserID)
	if val == nil {
		return uuid.UUID{}, fmt.Errorf("user_id not found in context")
	}

	userID, ok := val.(uuid.UUID)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("invalid user_id type in context")
	}

	return userID, nil
}
