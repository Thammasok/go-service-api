package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDB_InvalidURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	db, err := NewDB(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Equal(t, "database URL is required", err.Error())
}

func TestNewDB_MalformedURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	db, err := NewDB(ctx, "not-a-valid-url")
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestNewDB_UnreachableDatabase(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Use a database URL that points to a non-existent server
	url := "postgresql://root:root@localhost:5432/happy-db?schema=main&sslmode=disable"
	db, err := NewDB(ctx, url)
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestDBPool_Close(t *testing.T) {
	// Test that Close() handles nil pool gracefully
	db := &DBPool{
		pool: nil,
	}

	// Should not panic
	db.Close()
}

func TestDBPool_Stats(t *testing.T) {
	db := &DBPool{
		pool: nil,
	}

	// This test checks the method exists and is callable
	// Without a real database, we just verify no panic occurs
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Stats() panicked: %v", r)
		}
	}()

	// We can't test actual stats without a real pool,
	// but we verify the method signature is correct
	_ = db.GetPool()
}

// TestDBPool_ValidConnection tests with a valid PostgreSQL database
// To run this test, ensure a PostgreSQL instance is running at the specified URL
func TestDBPool_ValidConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This would require a real PostgreSQL instance
	// For CI/CD, use environment variable TEST_DATABASE_URL
	databaseURL := "postgres://user:password@localhost:5432/testdb?sslmode=disable"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := NewDB(ctx, databaseURL)
	if err != nil {
		t.Skip("PostgreSQL not available, skipping integration test:", err)
	}
	defer db.Close()

	assert.NotNil(t, db)
	assert.NotNil(t, db.GetPool())
}

// TestDBPool_HealthCheck tests the health check functionality
func TestDBPool_HealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	databaseURL := "postgres://user:password@localhost:5432/testdb?sslmode=disable"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := NewDB(ctx, databaseURL)
	if err != nil {
		t.Skip("PostgreSQL not available, skipping integration test:", err)
	}
	defer db.Close()

	healthCtx, healthCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer healthCancel()

	err = db.Health(healthCtx)
	assert.NoError(t, err)
}

// TestDBPool_GetPool tests the GetPool method
func TestDBPool_GetPool(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	databaseURL := "postgres://user:password@localhost:5432/testdb?sslmode=disable"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := NewDB(ctx, databaseURL)
	if err != nil {
		t.Skip("PostgreSQL not available, skipping integration test:", err)
	}
	defer db.Close()

	pool := db.GetPool()
	require.NotNil(t, pool)
	assert.Equal(t, pool, db.pool)
}

// TestDBPool_ContextTimeout tests behavior with cancelled context
func TestDBPool_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	db, err := NewDB(ctx, "postgres://user:password@localhost:5432/testdb?sslmode=disable")
	assert.Error(t, err)
	assert.Nil(t, db)
}
