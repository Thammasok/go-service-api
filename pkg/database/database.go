package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBPool is a wrapper around pgxpool for database operations
type DBPool struct {
	pool *pgxpool.Pool
}

// NewDB creates a new database connection pool with the given DSN
func NewDB(ctx context.Context, databaseURL string) (*DBPool, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	// Parse the connection string
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure connection pool settings
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConnIdleTime = 2 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DBPool{pool: pool}, nil
}

// GetPool returns the underlying pgxpool.Pool
func (db *DBPool) GetPool() *pgxpool.Pool {
	return db.pool
}

// Query executes a query and returns rows
func (db *DBPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return db.pool.Query(ctx, sql, args...)
}

// QueryRow executes a query that returns at most one row
func (db *DBPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return db.pool.QueryRow(ctx, sql, args...)
}

// Exec executes a command
func (db *DBPool) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.pool.Exec(ctx, sql, args...)
}

// Begin starts a new transaction
func (db *DBPool) Begin(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}

// Close closes all connections in the pool
func (db *DBPool) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// Health checks the health of the database connection
func (db *DBPool) Health(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

// Stats returns the current pool statistics
func (db *DBPool) Stats() pgxpool.Stat {
	return *db.pool.Stat()
}
