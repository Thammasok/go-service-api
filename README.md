# Go Service API

A high-performance Go REST API service using Fiber web framework and PostgreSQL database with pgx connection pooling.

## Project Structure

```
go-service-api/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   ├── config.go               # Configuration management
│   │   └── database.go             # Database connection setup
│   ├── domain/
│   │   ├── mod.go                  # Domain model initialization
│   │   ├── authentication/         # Auth domain
│   │   ├── common/                 # Common utilities
│   │   │   ├── router.go           # Request routing
│   │   │   ├── health/             # Health check handlers
│   │   │   └── home/               # Home endpoints
│   │   └── examples/               # Example domain
│   └── middleware/
│       └── error.go                # Error handling middleware
├── pkg/
│   ├── database/
│   │   ├── database.go             # Database pool wrapper
│   │   └── database_test.go        # Database tests
│   ├── logger/
│   │   ├── logger.go               # Logging utility
│   │   └── logger_test.go          # Logger tests
│   └── validate/                   # Validation utilities
├── migrations/
│   └── 202602181953_User.sql       # Database migrations
├── go.mod                          # Go module definition
├── Makefile                        # Build and run commands
├── ERROR_HANDLING.md               # Error handling documentation
└── README.md                       # This file
```

## Prerequisites

- Go 1.25.0 or higher
- PostgreSQL 12 or higher
- Make (optional, for convenient commands)

## Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd go-service-api
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   Create a `.env` file in the root directory:
   ```env
   PORT=8080
   ENV=development
   LOG_LEVEL=info
   DATABASE_URL=postgres://user:password@localhost:5432/go_service_db?sslmode=disable
   READ_TIMEOUT=5s
   WRITE_TIMEOUT=10s
   ```

4. **Set up PostgreSQL database**
   ```bash
   # Create database
   createdb go_service_db
   
   # Run migrations
   psql go_service_db < migrations/202602181953_User.sql
   ```

## Running the Application

### Development Mode
```bash
go run cmd/server/main.go
```

Or using Makefile:
```bash
make dev
```

### Production Build
```bash
go build -o bin/app cmd/server/main.go
./bin/app
```

Or using Makefile:
```bash
make build
make run
```

## Configuration

Configuration is loaded from environment variables with defaults:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | HTTP server port |
| `ENV` | development | Environment (development, staging, production) |
| `LOG_LEVEL` | info | Log level (debug, info, warn, error) |
| `DATABASE_URL` | (required) | PostgreSQL connection string |
| `READ_TIMEOUT` | 5s | HTTP server read timeout |
| `WRITE_TIMEOUT` | 10s | HTTP server write timeout |

### Loading Configuration

```go
// Load from .env file, fallback to environment variables
cfg, err := config.LoadFromFile(".env")
if err != nil {
    cfg = config.MustLoadFromEnv()
}
```

## Database Usage

### Initialize Database Connection

```go
import (
    "context"
    "dvith.com/go-service-api/pkg/database"
)

// Create connection pool
ctx := context.Background()
db, err := database.NewDB(ctx, "postgres://user:password@localhost:5432/dbname")
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

### Database Operations

#### Query (multiple rows)
```go
rows, err := db.Query(ctx, "SELECT id, email FROM users WHERE active = true")
if err != nil {
    log.Fatal(err)
}
defer rows.Close()

for rows.Next() {
    var id string
    var email string
    rows.Scan(&id, &email)
    // Process row
}
```

#### QueryRow (single row)
```go
var email string
row := db.QueryRow(ctx, "SELECT email FROM users WHERE id = $1", userID)
err := row.Scan(&email)
if err != nil {
    log.Fatal(err)
}
```

#### Execute (INSERT, UPDATE, DELETE)
```go
result, err := db.Exec(ctx, 
    "INSERT INTO users (email, password_hash) VALUES ($1, $2)", 
    email, 
    passwordHash,
)
if err != nil {
    log.Fatal(err)
}
rowsAffected := result.RowsAffected()
```

#### Transactions
```go
tx, err := db.Begin(ctx)
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback(ctx)

// Perform operations within transaction
_, err = tx.Exec(ctx, "INSERT INTO users ...")
if err != nil {
    log.Fatal(err)
}

// Commit transaction
err = tx.Commit(ctx)
```

### Database Pool Statistics

```go
stats := db.Stats()
// Access pool stats:
// stats.AcquiredConns()
// stats.IdleConns()
// stats.CreatedConns()
// stats.MaxConns
// stats.MinConns
```

### Health Check

```go
err := db.Health(ctx)
if err != nil {
    log.Println("Database is unavailable")
}
```

## API Endpoints

### Health Check
```
GET /api/health
```
Returns server health status and database connection status.

### Home
```
GET /
```
Returns a welcome message.

## Testing

### Run All Tests
```bash
go test ./...
```

### Run Tests Verbosely
```bash
go test ./... -v
```

### Run Tests with Coverage
```bash
go test ./... -cover
```

### Run Specific Package Tests
```bash
go test ./pkg/database -v
go test ./pkg/logger -v
```

### Test Results Summary
- **internal/middleware**: Error handling response tests
- **pkg/database**: Database connection pool tests
- **pkg/logger**: Logging functionality tests

## Database Schema

### users table
The main user table created by the migration:

```sql
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  full_name VARCHAR(255),
  username VARCHAR(100) UNIQUE,
  is_active BOOLEAN DEFAULT true,
  email_verified BOOLEAN DEFAULT false,
  verified_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  CONSTRAINT email_not_empty CHECK (email != '')
);

-- Indexes for performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
```

## Logging

The application uses Logrus for structured logging:

### Log Levels
- `debug`: Detailed debugging information
- `info`: General informational messages
- `warn`: Warning messages for potentially harmful situations
- `error`: Error messages for failures

### Using Logger
```go
import "dvith.com/go-service-api/pkg/logger"

// Simple logging
logger.Info("User created")
logger.Error("Database connection failed")

// Logging with fields
logger.Info("User created", map[string]any{
    "user_id": userID,
    "email": email,
})
```

## Error Handling

The application includes comprehensive error handling with structured error responses. See [ERROR_HANDLING.md](./ERROR_HANDLING.md) for detailed error handling documentation.

### Error Response Format
```json
{
  "code": 400,
  "message": "Bad Request",
  "status": "error"
}
```

## Connection Pool Configuration

The database connection pool is configured with the following settings:

| Setting | Value | Description |
|---------|-------|-------------|
| Max Connections | 25 | Maximum concurrent connections |
| Min Connections | 5 | Minimum idle connections |
| Max Conn Lifetime | 5 minutes | Maximum time a connection can be reused |
| Max Conn Idle Time | 2 minutes | Maximum idle time before closing |
| Health Check Period | 1 minute | How often to check connection health |

These settings can be customized in `pkg/database/database.go`.

## Development Guidelines

### Adding a New Endpoint
1. Create a handler in the appropriate domain folder
2. Define the route in `internal/domain/<feature>/routes.go`
3. Register the route in `internal/domain/mod.go`
4. Add tests for the handler

### Adding a Database Migration
1. Create a new migration file in `migrations/` with timestamp prefix
2. Write SQL in the file
3. Run the migration locally to test
4. Commit the migration file

### Code Style
- Use `camelCase` for variable and function names
- Use `PascalCase` for type and constant names
- Include error handling for all I/O operations
- Write tests for new functionality

## Common Issues

### Connection Refused Error
**Problem**: `dial error: dial tcp 127.0.0.1:5432: connect: connection refused`

**Solution**: 
- Ensure PostgreSQL is running: `brew services start postgresql` (macOS)
- Check database URL is correct in `.env`
- Verify PostgreSQL is listening on the correct port

### Authentication Failed Error
**Problem**: `FATAL: password authentication failed for user`

**Solution**:
- Verify username and password in `DATABASE_URL`
- Check PostgreSQL user exists: `psql -U postgres -c "SELECT * FROM pg_user;"`
- Reset password if needed: `ALTER USER username WITH PASSWORD 'newpassword';`

### Database Does Not Exist
**Problem**: `database "go_service_db" does not exist`

**Solution**:
- Create database: `createdb go_service_db`
- Or modify `DATABASE_URL` to use existing database

## Performance Tips

1. **Connection Pool**: The pool maintains 5-25 connections. Adjust based on load.
2. **Indexes**: Ensure indexes exist on frequently queried columns.
3. **Prepared Statements**: Use parameterized queries (`$1`, `$2`) to prevent SQL injection.
4. **Timeouts**: Set appropriate context timeouts for database operations.

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues, questions, or contributions, please open an issue in the repository.
