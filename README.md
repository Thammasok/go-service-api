# Go Service API

A high-performance Go REST API service using Fiber web framework and PostgreSQL database with pgx connection pooling.

## Project Structure

```
go-service-api/
├── cmd/
│   └── server/
│       └── main.go                      # Application entry point
├── internal/
│   ├── config/
│   │   ├── config.go                    # Configuration management
│   │   └── database.go                  # Database connection setup
│   ├── domain/
│   │   ├── mod.go                       # Domain model initialization
│   │   ├── authentication/              # Auth domain
│   │   │   ├── auth_route.go                # Route registration
│   │   │   └── signup/                      # User signup
│   │   │       ├── signup_handler.go        # HTTP handler
│   │   │       ├── signup_service.go        # Business logic
│   │   │       ├── signup_repository.go     # Data access
│   │   │       └── validator.go            # Request validation
│   │   ├── common/                      # Common utilities
│   │   │   ├── common_route.go          # Request routing
│   │   │   ├── health/                  # Health check handlers
│   │   │   └── home/                    # Home endpoints
│   │   └── examples/                    # Example domain
│   ├── middleware/
│   │   └── error.go                     # Error handling middleware
│   └── utils/
│       └── hash_password/               # Password hashing utility
│           ├── hash_password.go         # Argon2 hashing
│           └── hash_password_test.go    # Tests & benchmarks
├── pkg/
│   ├── database/
│   │   ├── database.go                  # Database pool wrapper
│   │   └── database_test.go             # Database tests
│   ├── logger/
│   │   ├── logger.go                    # Logging utility
│   │   └── logger_test.go               # Logger tests
│   └── validate/                        # Validation utilities
├── migrations/
│   └── 202602181953_User.sql            # Database migrations
├── go.mod                               # Go module definition
├── Makefile                             # Build and run commands
├── ARCHITECTURE.md                      # Architecture documentation
├── ERROR_HANDLING.md                    # Error handling documentation
└── README.md                            # This file
```

## Prerequisites

- Go 1.25.0 or higher
- PostgreSQL 12 or higher
- Make (optional, for convenient commands)

## Dependencies

### Core Dependencies

- **Fiber v3**: Web framework for HTTP handling ([github.com/gofiber/fiber/v3](https://github.com/gofiber/fiber/v3))
- **pgx v5**: PostgreSQL driver with connection pooling ([github.com/jackc/pgx/v5](https://github.com/jackc/pgx/v5))
- **Logrus**: Structured logging ([github.com/sirupsen/logrus](https://github.com/sirupsen/logrus))

### Utility Dependencies

- **Argon2**: Password hashing ([golang.org/x/crypto/argon2](https://pkg.go.dev/golang.org/x/crypto/argon2))
- **go-playground/validator**: Input validation ([github.com/go-playground/validator/v10](https://github.com/go-playground/validator))
- **UUID**: Unique identifier generation ([github.com/google/uuid](https://github.com/google/uuid))

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

| Variable        | Default       | Description                                                           |
| --------------- | ------------- | --------------------------------------------------------------------- |
| `PORT`          | 8080          | HTTP server port                                                      |
| `ENV`           | `development` | Environment (`development`, `local`, `staging`, `test`, `production`) |
| `DATABASE_URL`  | (required)    | PostgreSQL connection string                                          |
| `READ_TIMEOUT`  | 5s            | HTTP server read timeout                                              |
| `WRITE_TIMEOUT` | 10s           | HTTP server write timeout                                             |

> **Log level is derived from `ENV` automatically.** See the [Logging](#logging) section for the mapping.

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
    "INSERT INTO users (email, password) VALUES ($1, $2)",
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

### User Signup

```
POST /api/v1/signup
```

Register a new user with email, password, username, and full name.

**Request Body**:

```json
{
  "email": "user@example.com",
  "password": "securePassword123",
  "username": "john_doe",
  "full_name": "John Doe"
}
```

**Validation Rules**:

- Email: required, valid email format
- Password: required, minimum 8 characters
- Username: required, 3-100 characters
- Full Name: required, maximum 255 characters

**Success Response (201 Created)**:

```json
{
  "message": "User registered successfully",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "username": "john_doe",
    "fullName": "John Doe",
    "isActive": true,
    "createdAt": "2026-02-18T10:30:00Z"
  }
}
```

**Error Response (400 Bad Request)**:

```json
{
  "error": "Validation failed",
  "errors": [
    {
      "field": "Email",
      "message": "Email must be a valid email address"
    },
    {
      "field": "Password",
      "message": "Password must be at least 8 characters"
    }
  ]
}
```

### Health Check

```
GET /api/v1/health
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
# Database tests
go test ./pkg/database -v

# Logger tests
go test ./pkg/logger -v

# Password hashing tests
go test ./internal/utils/hash_password -v
```

### Hash Password Tests

Comprehensive tests for password hashing and verification:

```bash
# Run all hash password tests
go test ./internal/utils/hash_password -v

# Run with benchmarks
go test ./internal/utils/hash_password -bench=. -benchmem

# Run specific test
go test ./internal/utils/hash_password -run TestHashPassword -v
```

**Performance Benchmarks** (Apple M5):

- HashPassword: ~27.5ms per operation
- CheckPassword: ~26.4ms per operation

### Test Results Summary

- **internal/middleware**: Error handling response tests
- **internal/utils/hash_password**: Argon2 hashing and verification tests
- **pkg/database**: Database connection pool tests
- **pkg/logger**: Logging functionality tests

## Database Schema

### users table

The main user table created by the migration:

```sql
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) UNIQUE NOT NULL,
  password VARCHAR(255) NOT NULL,
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

## Password Security

### Hashing Algorithm: Argon2-ID

Passwords are securely hashed using Argon2-ID (OWASP recommended):

```go
import "dvith.com/go-service-api/internal/utils/hash_password"

// Hash a password
hashedPassword, err := hash_password.HashPassword("user_password")

// Verify a password
isValid := hash_password.CheckPassword("user_password", hashedPassword)
```

### Configuration

- **Algorithm**: Argon2-ID (resistant to GPU and side-channel attacks)
- **Time Cost**: 3 iterations
- **Memory Cost**: 64 MB
- **Parallelism**: 4 threads
- **Output Length**: 32 bytes (256-bit hash)

### Security Features

- Passwords never stored in plain text
- Hashed passwords excluded from API responses
- Minimum 8 character password requirement
- Passwords automatically hashed during user signup

## Input Validation

The application uses `go-playground/validator/v10` for robust input validation:

### Signup Request Validation

```go
type SignupRequest struct {
    Email    string `validate:"required,email"`           // Required, valid email
    Password string `validate:"required,min=8"`           // Required, min 8 chars
    FullName string `validate:"required,max=255"`         // Required, max 255 chars
    Username string `validate:"required,min=3,max=100"` // Required, 3-100 chars
}
```

### Validation Error Response

```json
{
  "error": "Validation failed",
  "errors": [
    {
      "field": "Password",
      "message": "Password must be at least 8 characters"
    }
  ]
}
```

## Logging

The application uses Logrus for structured logging. Log level is automatically determined by `ENV` — no manual `LOG_LEVEL` setting required.

### Log Level by Environment

| `ENV`                  | Log Level | Output                      |
| ---------------------- | --------- | --------------------------- |
| `development`, `local` | DEBUG     | All logs, with ANSI colours |
| `staging`, `test`      | INFO      | Info, Warn, Error           |
| `production`           | ERROR     | Errors only                 |
| (other)                | INFO      | Info, Warn, Error           |

### Using Logger

```go
import "dvith.com/go-service-api/pkg/logger"

// Logging with fields
logger.Info("server started", map[string]any{"port": 8080})
logger.Warn("slow query",     map[string]any{"ms": 350})
logger.Error("db error",      map[string]any{"err": err})
logger.Debug("parsed body",   map[string]any{"body": body})
```

### Initialise from Environment

```go
// Called once at startup — sets level and formatter automatically
logger.InitFromEnv(cfg.Env)
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

### Validation Errors

Validation errors provide field-level details for debugging:

```json
{
  "error": "Validation failed",
  "errors": [
    {
      "field": "Email",
      "message": "Email must be a valid email address"
    },
    {
      "field": "Password",
      "message": "Password must be at least 8 characters"
    }
  ]
}
```

## Connection Pool Configuration

The database connection pool is configured with the following settings:

| Setting             | Value     | Description                             |
| ------------------- | --------- | --------------------------------------- |
| Max Connections     | 25        | Maximum concurrent connections          |
| Min Connections     | 5         | Minimum idle connections                |
| Max Conn Lifetime   | 5 minutes | Maximum time a connection can be reused |
| Max Conn Idle Time  | 2 minutes | Maximum idle time before closing        |
| Health Check Period | 1 minute  | How often to check connection health    |

These settings can be customized in `pkg/database/database.go`.

## Authentication & User Signup

### Signup Flow

The signup process handles user registration with secure password hashing:

1. **Request Validation**: User input is validated using `go-playground/validator`
   - Email format validation
   - Password strength requirements (min 8 characters)
   - Username and full name constraints

2. **Password Hashing**: Passwords are hashed using Argon2-ID
   - Memory-hard algorithm resistant to GPU attacks
   - Never stored in plain text
   - Configuration: 3 iterations, 64MB memory, 4 parallelism

3. **User Persistence**: User data is saved to PostgreSQL database
   - UUID auto-generated for user ID
   - Timestamps automatically set
   - Unique email and username constraints enforced

4. **Response**: User object returned (without password hash)

**Example Signup Request**:

```bash
curl -X POST http://localhost:8080/api/v1/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePassword123",
    "username": "john_doe",
    "full_name": "John Doe"
  }'
```

**Example Success Response (201 Created)**:

```json
{
  "message": "User registered successfully",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "john@example.com",
    "username": "john_doe",
    "fullName": "John Doe",
    "isActive": true,
    "createdAt": "2026-02-18T10:30:00Z"
  }
}
```

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

## Documentation

- **[ARCHITECTURE.md](./doc/ARCHITECTURE.md)** - Detailed system architecture, design patterns, and layered architecture documentation
- **[ERROR_HANDLING.md](./doc/ERROR_HANDLING.md)** - Error handling patterns and response formats
- **[DATABASE.md](./doc/DATABASE.md)** - Database connection pool setup, query patterns, transactions, and migrations
- **[PASSWORD_STRENGTH.md](./doc/PASSWORD_STRENGTH.md)** - Password strength requirements and validation rules
