package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	// URL is the base URL for the service (optional, used for generating links).
	URL string

	// Port the HTTP server will listen on.
	Port int

	// Env application environment, e.g. development, staging, production
	Env string

	// LogLevel textual log level (debug, info, warn, error)
	LogLevel string

	// Database connection string (optional)
	DatabaseURL string

	// ReadTimeout for HTTP server
	ReadTimeout time.Duration

	// WriteTimeout for HTTP server
	WriteTimeout time.Duration
}

// LoadFromEnv reads configuration from environment variables, applying defaults
// and returning an error if parsing fails for any provided value.
func LoadFromEnv() (Config, error) {
	var c Config

	// defaults
	c.Port = 8080
	c.Env = "development"
	c.LogLevel = "info"
	c.DatabaseURL = ""
	c.ReadTimeout = 5 * time.Second
	c.WriteTimeout = 10 * time.Second

	if v := os.Getenv("PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return c, fmt.Errorf("invalid PORT: %w", err)
		}
		c.Port = p
	}
	if v := os.Getenv("ENV"); v != "" {
		c.Env = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		c.LogLevel = v
	}
	if v := os.Getenv("DATABASE_URL"); v != "" {
		c.DatabaseURL = v
	}

	return c, nil
}

// MustLoadFromEnv loads config and panics on error. Useful for simple main functions.
func MustLoadFromEnv() Config {
	cfg, err := LoadFromEnv()
	if err != nil {
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
