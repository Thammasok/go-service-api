package private

import (
	"dvith.com/go-service-api/internal/middleware"
	"dvith.com/go-service-api/pkg/database"
	"dvith.com/go-service-api/pkg/logger"
	"github.com/gofiber/fiber/v3"
)

// ProfileResponse represents the user profile response
type ProfileResponse struct {
	UserID string `json:"user_id"`
	// Add more fields as needed
}

// ProfileHandler retrieves the authenticated user's profile
func ProfileHandler(db *database.DBPool) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get user ID from context (set by AuthMiddleware)
		userID, err := middleware.GetUserIDFromContext(c)
		if err != nil {
			logger.Warn("failed to get user id from context", map[string]any{
				"error": err.Error(),
			})
			return middleware.AuthErrorResponse(c, "user not authenticated")
		}

		logger.Debug("fetching user profile", map[string]any{
			"user_id": userID.String(),
		})

		// TODO: Query database for user profile
		// For now, return the user ID

		return c.Status(fiber.StatusOK).JSON(ProfileResponse{
			UserID: userID.String(),
		})
	}
}
