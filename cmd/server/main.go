package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dvith.com/go-service-api/internal/config"
	"dvith.com/go-service-api/internal/domain"
	"dvith.com/go-service-api/pkg/database"
	"dvith.com/go-service-api/pkg/logger"
	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	// Prefer loading configuration from a local .env-like file into a
	// Config object. If the file isn't present or fails to parse, fall
	// back to reading from the process environment.
	cfg, err := config.LoadFromFile(".env")
	if err != nil {
		logger.Warn("could not load configuration from .env file", map[string]any{"error": err.Error()})
		cfg = config.MustLoadFromEnv()
	}

	logger.InitFromEnv(cfg.Env)

	// If a database URL is provided, initialize the connection pool.
	db, err := database.NewDB(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to initialize database", map[string]any{"error": err.Error()})
	}
	defer db.Close()

	// set up routes and start the server
	domain.Init(app, db)

	addr := fmt.Sprintf(":%d", cfg.Port)

	// Start server in background so we can handle graceful shutdown.
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- app.Listen(addr)
	}()

	// trap signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Info("shutdown signal received", map[string]any{"signal": sig.String()})

		// give the server up to 10s to shut down gracefully
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		done := make(chan struct{})
		go func() {
			if err := app.Shutdown(); err != nil {
				logger.Error("error during shutdown", map[string]any{"err": err.Error()})
			}
			close(done)
		}()

		select {
		case <-done:
			logger.Info("server stopped", nil)
		case <-ctx.Done():
			logger.Warn("graceful shutdown timed out", nil)
		}

	case err := <-srvErr:
		if err != nil {
			logger.Error("server listen error", map[string]any{"err": err.Error(), "addr": addr})
			os.Exit(1)
		}
	}
}
