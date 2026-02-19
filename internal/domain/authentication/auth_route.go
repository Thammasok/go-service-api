package authentication

import (
	"dvith.com/go-service-api/internal/config"
	refreshtoken "dvith.com/go-service-api/internal/domain/authentication/refresh_token"
	"dvith.com/go-service-api/internal/domain/authentication/signin"
	"dvith.com/go-service-api/internal/domain/authentication/signup"
	"dvith.com/go-service-api/pkg/database"
	"github.com/gofiber/fiber/v3"
)

func Routers(app fiber.Router, db *database.DBPool, cfg config.Config) {
	// Authentication routes
	app.Post("/auth/signup", signup.SignupHandler(db, cfg))
	app.Post("/auth/signin", signin.SigninHandler(db, cfg))
	app.Post("/auth/refresh-token", refreshtoken.RefreshTokenHandler(db, cfg))
}
