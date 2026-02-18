# Database Documentation

## Overview

The database layer uses PostgreSQL with `jackc/pgx/v5` for high-performance connections and connection pooling. This document provides detailed information about database setup, usage, and best practices.

## Connection Pool

### Architecture

The database connection pool is implemented in `pkg/database/database.go` and provides:

- **Connection Pooling**: Manages a pool of PostgreSQL connections
- **Health Checks**: Periodic health checks ensure connections are alive
- **Graceful Shutdown**: Properly closes all connections on application exit
- **Error Recovery**: Handles connection failures and timeouts
- **Statistics**: Provides pool statistics for monitoring

### Configuration

```go
config.MaxConns = 25              // Maximum concurrent connections
config.MinConns = 5               // Minimum idle connections
config.MaxConnLifetime = 5m       // Reuse connections for max 5 minutes
config.MaxConnIdleTime = 2m       // Close idle connections after 2 minutes
config.HealthCheckPeriod = 1m     // Check connection health every minute
```

### Initialize Connection Pool

```go
package main

import (
    "context"
    "dvith.com/go-service-api/pkg/database"
)

func main() {
    ctx := context.Background()
    
    // Create the connection pool
    db, err := database.NewDB(ctx, "postgres://user:pass@localhost:5432/dbname")
    if err != nil {
        panic(err)
    }
    defer db.Close()
    
    // Use db for queries
    // ...
}
```

## Basic Operations

### Query Multiple Rows

```go
rows, err := db.Query(ctx, "SELECT id, email, created_at FROM users WHERE is_active = true")
if err != nil {
    log.Fatal(err)
}
defer rows.Close()

for rows.Next() {
    var id string
    var email string
    var createdAt time.Time
    
    if err := rows.Scan(&id, &email, &createdAt); err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("ID: %s, Email: %s, Created: %s\n", id, email, createdAt)
}

if err = rows.Err(); err != nil {
    log.Fatal(err)
}
```

### Query Single Row

```go
var (
    id    string
    email string
)

row := db.QueryRow(ctx, "SELECT id, email FROM users WHERE id = $1", userID)
if err := row.Scan(&id, &email); err != nil {
    if err == pgx.ErrNoRows {
        // User not found
        return fmt.Errorf("user not found")
    }
    return err
}
```

### Insert

```go
var id string
err := db.QueryRow(ctx,
    "INSERT INTO users (email, password_hash, full_name) VALUES ($1, $2, $3) RETURNING id",
    email,
    passwordHash,
    fullName,
).Scan(&id)

if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

fmt.Printf("Created user with ID: %s\n", id)
```

### Update

```go
result, err := db.Exec(ctx,
    "UPDATE users SET email = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
    newEmail,
    userID,
)

if err != nil {
    return err
}

// Check rows affected
fmt.Printf("Rows updated: %d\n", result.RowsAffected())
```

### Delete

```go
result, err := db.Exec(ctx,
    "DELETE FROM users WHERE id = $1",
    userID,
)

if err != nil {
    return err
}

if result.RowsAffected() == 0 {
    return fmt.Errorf("user not found")
}
```

## Transactions

Transactions are essential for maintaining data consistency:

```go
// Start transaction
tx, err := db.Begin(ctx)
if err != nil {
    return err
}

// Use defer to ensure rollback on error
defer tx.Rollback(ctx)

// Perform operations
result, err := tx.Exec(ctx, "INSERT INTO users (email) VALUES ($1)", "user@example.com")
if err != nil {
    return err // Automatically rolled back
}

// Another operation
_, err = tx.Exec(ctx, "UPDATE stats SET user_count = user_count + 1")
if err != nil {
    return err // Automatically rolled back
}

// Commit if all succeeded
if err = tx.Commit(ctx); err != nil {
    return err
}
```

### Savepoints (Nested Transactions)

```go
tx, err := db.Begin(ctx)
if err != nil {
    return err
}
defer tx.Rollback(ctx)

// Create savepoint
sp, err := tx.SavePoint(ctx, "sp1")
if err != nil {
    return err
}

// Do some work
_, err = tx.Exec(ctx, "INSERT INTO users (email) VALUES ($1)", "user@example.com")
if err != nil {
    // Rollback to savepoint
    sp.Rollback(ctx)
    // But continue with different operation
}

// Commit everything
tx.Commit(ctx)
```

## Context and Timeouts

Always use contexts with timeouts for database operations:

```go
// Query with 5 second timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

rows, err := db.Query(ctx, "SELECT * FROM users")
```

## Error Handling

### Common Errors

```go
import "github.com/jackc/pgx/v5"

// Handle no rows error
row := db.QueryRow(ctx, "SELECT * FROM users WHERE id = $1", id)
err := row.Scan(&u)
if err == pgx.ErrNoRows {
    // User not found
} else if err != nil {
    // Other error
}

// Check for constraint violations
var pgErr *pgconn.PgError
if errors.As(err, &pgErr) {
    switch pgErr.Code {
    case "23505": // Unique constraint violation
        // Handle duplicate
    case "23503": // Foreign key constraint violation
        // Handle referenced record not found
    }
}
```

## Performance Best Practices

### 1. Use Parameterized Queries

**Bad** (SQL injection risk):
```go
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)
db.Query(ctx, query)
```

**Good** (Safe):
```go
db.Query(ctx, "SELECT * FROM users WHERE email = $1", email)
```

### 2. Index Frequently Queried Columns

```sql
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_created_at ON users(created_at);
```

### 3. Batch Operations

For bulk inserts/updates, use batch operations:

```go
batch := &pgx.Batch{}

for _, user := range users {
    batch.Queue("INSERT INTO users (email, password_hash) VALUES ($1, $2)",
        user.Email, user.PasswordHash)
}

results := db.GetPool().SendBatch(ctx, batch)
defer results.Close()
```

### 4. Use Connection Pool Efficiently

- Set appropriate pool sizes based on workload
- Monitor pool statistics
- Avoid leaving contexts open too long

### 5. Query Complex Data Efficiently

```go
// Use DISTINCT to avoid duplicates
rows, _ := db.Query(ctx, "SELECT DISTINCT email FROM users")

// Use LIMIT to prevent large result sets
rows, _ := db.Query(ctx, "SELECT * FROM users LIMIT 1000")

// Use aggregate functions instead of fetching all rows
row := db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE is_active = true")
```

## Pool Statistics

Monitor connection pool health:

```go
stats := db.Stats()
fmt.Printf("Acquired: %d, Idle: %d, Created: %d\n",
    stats.AcquiredConns(),
    stats.IdleConns(),
    stats.CreatedConns(),
)
```

## Health Checks

Verify database connectivity:

```go
// Check connection health
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

if err := db.Health(ctx); err != nil {
    log.Printf("Database unhealthy: %v\n", err)
}
```

## Migrations

Database migrations are stored in `migrations/` directory with timestamp prefixes.

### Running Migrations

```bash
# Using psql
psql -h localhost -U user -d dbname -f migrations/202602181953_User.sql

# Or programmatically
rows, _ := os.ReadFile("migrations/202602181953_User.sql")
db.Exec(context.Background(), string(rows))
```

### Creating Migrations

Create a new file with format: `YYYYMMDDHHMM_Description.sql`

```sql
-- 202602181953_User.sql
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
```

## Testing

Database tests use the actual PostgreSQL connection or integration tests:

### Unit Tests (No Database Required)
```go
// Test error handling without a real database
func TestNewDB_InvalidURL(t *testing.T) {
    db, err := database.NewDB(context.Background(), "")
    assert.Error(t, err)
    assert.Nil(t, db)
}
```

### Integration Tests (Requires PostgreSQL)
```bash
# Run all tests
go test ./... -v

# Run only integration tests
go test ./... -v -run Integration

# Skip long tests
go test ./... -short
```

## Troubleshooting

### Connection Issues

1. **Connection Refused**
   ```
   Error: failed to connect to `localhost:5432`: dial tcp 127.0.0.1:5432: connect: connection refused
   ```
   Solution: Ensure PostgreSQL service is running

2. **Authentication Failed**
   ```
   Error: FATAL: password authentication failed for user "user"
   ```
   Solution: Check credentials in DATABASE_URL

3. **Database Does Not Exist**
   ```
   Error: database "go_service_db" does not exist
   ```
   Solution: Create the database with `createdb`

### Performance Issues

1. **Slow Queries**: Check if indexes exist on WHERE clause columns
2. **Connection Pool Exhausted**: Increase `MaxConns` or reduce context timeout
3. **High Memory Usage**: Reduce `MaxConns` or check for connection leaks

## Advanced Features

### Server-Side Cursors

For very large result sets:

```go
// Use a cursor instead of loading all rows
rows, err := db.Query(ctx, "DECLARE cur CURSOR FOR SELECT * FROM large_table")
```

### Prepared Statements

```go
stmt, err := db.GetPool().Prepare(ctx, "insert_user", 
    "INSERT INTO users (email) VALUES ($1)")
if err != nil {
    return err
}

// Reuse the prepared statement
_, err = db.GetPool().Exec(ctx, "insert_user", "user@example.com")
```

### Connection Pooling Monitoring

Enable logging to monitor pool activity:

```go
config := pgxpool.Config{
    ConnConfig: pgx.ConnConfig{
        Tracer: &pgx.LogTracer{
            Logger:   logger.Info,
            LogLevel: pgx.LogLevelDebug,
        },
    },
}
```

## Resources

- [jackc/pgx Documentation](https://github.com/jackc/pgx)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Connection Pooling Best Practices](https://wiki.postgresql.org/wiki/Number_Of_Database_Connections)
