package authentication

import (
	"dvith.com/go-service-api/internal/config"
	"dvith.com/go-service-api/internal/domain/authentication/signin"
	"dvith.com/go-service-api/internal/domain/authentication/signup"
	"dvith.com/go-service-api/pkg/database"
	"github.com/gofiber/fiber/v3"
)

func Routers(app fiber.Router, db *database.DBPool, cfg config.Config) {
	// Authentication routes
	app.Post("/signup", signup.SignupHandler(db, cfg))
	app.Post("/signin", signin.SigninHandler(db, cfg))
}
