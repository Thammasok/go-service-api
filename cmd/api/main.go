package main

import (
	"fmt"
	"os"
	"strings"

	"dvith.com/go-service-api/internal/config"
	"dvith.com/go-service-api/internal/routes"
	"dvith.com/go-service-api/pkg"
	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	// Prefer loading configuration from a local .env-like file into a
	// Config object. If the file isn't present or fails to parse, fall
	// back to reading from the process environment.
	cfg, err := config.LoadFromFile(".env")
	if err != nil {
		cfg = config.MustLoadFromEnv()
	}

	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		pkg.SetLevel(pkg.DebugLevel)
	case "warn", "warning":
		pkg.SetLevel(pkg.WarnLevel)
	case "error":
		pkg.SetLevel(pkg.ErrorLevel)
	default:
		pkg.SetLevel(pkg.InfoLevel)
	}

	// Log the active log level and port so it's visible on startup.
	pkg.Info("starting service", map[string]any{"level": strings.ToLower(cfg.LogLevel), "port": cfg.Port})

	// set up routes and start the server
	routes.SetupRoutes(app)

	addr := fmt.Sprintf(":%d", cfg.Port)
	if err := app.Listen(addr); err != nil {
		pkg.Error("failed to start server", map[string]any{"err": err.Error(), "addr": addr})
		os.Exit(1)
	}
}
