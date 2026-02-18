package config

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	envconfig "github.com/sethvargo/go-envconfig"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	// URL is the base URL for the service (optional, used for generating links).
	URL string `env:"URL"`

	// Port the HTTP server will listen on.
	Port int `env:"PORT,default=8080"`

	// Env application environment, e.g. development, staging, production
	Env string `env:"ENV,default=development"`

	// LogLevel textual log level (debug, info, warn, error)
	LogLevel string `env:"LOG_LEVEL,default=info"`

	// Database connection string (optional)
	DatabaseURL string `env:"DATABASE_URL"`

	// ReadTimeout for HTTP server
	ReadTimeout time.Duration `env:"READ_TIMEOUT,default=5s"`

	// WriteTimeout for HTTP server
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT,default=10s"`
}

// LoadFromEnv loads configuration from environment variables using go-envconfig.
func LoadFromEnv() (Config, error) {
	var c Config
	if err := envconfig.Process(context.Background(), &c); err != nil {
		return c, fmt.Errorf("failed to load environment: %w", err)
	}
	return c, nil
}

// MustLoadFromEnv loads config and panics on error. Useful for simple main functions.
func MustLoadFromEnv() Config {
	cfg, err := LoadFromEnv()
	if err != nil {
		panic(err)
	}
	if err := cfg.Validate(); err != nil {
		panic(err)
	}
	return cfg
}

// LoadFromFile parses a simple KEY=VALUE file (like .env) into a Config.
// It does not modify process environment.
func LoadFromFile(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()

	vals := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}
		vals[key] = val
	}
	if err := scanner.Err(); err != nil {
		return Config{}, err
	}

	// Start with defaults then override from vals map.
	c := Config{
		Port:         8080,
		Env:          "development",
		LogLevel:     "info",
		DatabaseURL:  "",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if v, ok := vals["PORT"]; ok && v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return c, fmt.Errorf("invalid PORT in file: %w", err)
		}
		c.Port = p
	}
	if v, ok := vals["ENV"]; ok && v != "" {
		c.Env = v
	}
	if v, ok := vals["LOG_LEVEL"]; ok && v != "" {
		c.LogLevel = v
	}
	if v, ok := vals["DATABASE_URL"]; ok && v != "" {
		c.DatabaseURL = v
	}
	if v, ok := vals["URL"]; ok && v != "" {
		c.URL = v
	}
	if v, ok := vals["READ_TIMEOUT"]; ok && v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return c, fmt.Errorf("invalid READ_TIMEOUT in file: %w", err)
		}
		c.ReadTimeout = d
	}
	if v, ok := vals["WRITE_TIMEOUT"]; ok && v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return c, fmt.Errorf("invalid WRITE_TIMEOUT in file: %w", err)
		}
		c.WriteTimeout = d
	}

	return c, nil
}

// Validate checks that required configuration values are present and well-formed.
// It returns an error describing the first validation failure encountered.
func (c Config) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("PORT must be between 1 and 65535, got %d", c.Port)
	}

	env := strings.ToLower(c.Env)
	switch env {
	case "development", "staging", "production", "test", "local":
	default:
		return fmt.Errorf("ENV must be one of development|staging|production|test|local, got %q", c.Env)
	}

	lvl := strings.ToLower(c.LogLevel)
	switch lvl {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("LOG_LEVEL must be one of debug|info|warn|error, got %q", c.LogLevel)
	}

	if c.ReadTimeout <= 0 {
		return fmt.Errorf("READ_TIMEOUT must be > 0")
	}
	if c.WriteTimeout <= 0 {
		return fmt.Errorf("WRITE_TIMEOUT must be > 0")
	}

	if strings.ToLower(c.Env) == "production" && strings.TrimSpace(c.DatabaseURL) == "" {
		return fmt.Errorf("DATABASE_URL is required in production environment")
	}

	return nil
}
