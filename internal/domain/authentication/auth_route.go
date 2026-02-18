package authentication

import (
	"dvith.com/go-service-api/internal/domain/authentication/signup"
	"dvith.com/go-service-api/pkg/database"
	"github.com/gofiber/fiber/v3"
)

func Routers(app fiber.Router, db *database.DBPool) {
	// Signup routes
	app.Post("/signup", signup.SignupHandler(db))
}
