package private

import (
	"dvith.com/go-service-api/internal/config"
	"dvith.com/go-service-api/internal/middleware"
	"dvith.com/go-service-api/internal/security/token"
	"dvith.com/go-service-api/pkg/database"
	"github.com/gofiber/fiber/v3"
)

func Routers(app fiber.Router, db *database.DBPool, cfg config.Config) {
	// Initialize token manager for protected routes
	tm := token.NewTokenManager(token.TokenConfig{
		SecretKey:       cfg.JWTSecretKey,
		ExpirationTime:  cfg.JWTExpirationTime,
		RefreshDuration: cfg.JWTRefreshDuration,
		Issuer:          cfg.JWTIssuer,
	})

	// Create a group for protected routes that require authentication
	withAuth := app.Group("/user", middleware.AuthMiddleware(tm))

	// Protected routes (require valid access token)
	withAuth.Get("/profile", ProfileHandler(db))
	// Add more protected routes here as needed
}
