# Architecture Documentation

## System Architecture Overview

This Go Service API follows a layered architecture pattern designed for maintainability, scalability, and testability.

```
┌─────────────────────────────────────────────────────────┐
│                   HTTP Client / Browser                 │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│                    Fiber Web Framework                  │
│              (HTTP Server & Routing)                    │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│                     Middleware Layer                    │
│    (Error Handling, Logging, Authentication)           │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│                    Handler/Endpoint Layer               │
│         (Request Processing & Response Format)         │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│                   Business Logic Layer                  │
│         (Service, Domain Models, Rule Engine)          │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│                   Data Access Layer                     │
│            (Database Queries, Repositories)            │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│                 PostgreSQL Database                     │
│          (pgx Connection Pool & Transactions)          │
└─────────────────────────────────────────────────────────┘
```

## Directory Structure and Responsibilities

### `/cmd/server`

**Purpose**: Application entry point and initialization

**Contains**:

- `main.go`: Application startup logic
  - Configuration loading
  - Database connection setup
  - HTTP server initialization
  - Graceful shutdown handling

**Key Responsibilities**:

- Parse command-line arguments
- Load environment configuration
- Initialize all services
- Start the server
- Handle shutdown signals

### `/internal`

**Purpose**: Private application code (not exported)

#### `/internal/config`

**Purpose**: Configuration management and initialization

**Contains**:

- `config.go`: Configuration loading from environment variables
- `database.go`: Database connection setup

**Key Classes/Functions**:

- `Config`: Application configuration struct
- `LoadFromFile()`: Load config from `.env` file
- `LoadFromEnv()`: Load config from environment
- `MustLoadFromEnv()`: Load config with panic on error
- `NewDB()`: Initialize database connection pool
- `DBPool`: Database connection wrapper

### `/internal/domain`

**Purpose**: Business domain logic and endpoints

**Contains**:

- `mod.go`: Domain module initialization and registration
- `authentication/`: User authentication endpoints
  - `auth_route.go`: Authentication routes registration
  - `signup/`: User registration/signup functionality
    - `signup_handler.go`: HTTP handler for signup requests
    - `signup_service.go`: Business logic for user registration
    - `signup_repository.go`: Data access layer for user persistence
    - `validator.go`: Signup request validation using go-playground/validator
- `common/`: Common utilities and handlers
  - `common_route.go`: Request routing configuration
  - `health/`: Health check endpoints
  - `home/`: Home/welcome endpoints
- `examples/`: Example endpoints and routes

**Design Pattern**: Domain-Driven Design with Service-Repository Pattern

- Each domain has clear responsibility
- Routes registered in domain modules
- Handlers process specific domain requests
- Services contain business logic
- Repositories handle data persistence

**Signup Flow**:

```
POST /api/v1/signup
  ↓
SignupHandler -> Bind & Validate
  ↓
Validation with go-playground/validator
  ├─ Email required, valid format
  ├─ Password required, min 8 chars
  ├─ FullName required, max 255 chars
  └─ Username required, 3-100 chars
  ↓
SignupService.RegisterUser()
  ├─ Hash password with Argon2
  └─ Return user from database
  ↓
HTTP 201 Created with User Data
```

### `/internal/middleware`

**Purpose**: Cross-cutting concerns

**Contains**:

- `error.go`: Global error handling and response formatting

**Key Middleware**:

- Error handling with standard response format
- Logging of errors
- HTTP status code mapping

### `/pkg`

**Purpose**: Reusable packages (can be exported)

#### `/pkg/database`

**Purpose**: Database abstraction and connection pooling

**Contains**:

- `database.go`: Database pool wrapper
- `database_test.go`: Unit tests

**Key Classes/Functions**:

- `DBPool`: PostgreSQL connection pool wrapper
- `NewDB()`: Initialize connection pool
- `Query()`: Execute query returning multiple rows
- `QueryRow()`: Execute query returning single row
- `Exec()`: Execute command (INSERT, UPDATE, DELETE)
- `Begin()`: Start transaction
- `Close()`: Close connection pool
- `Health()`: Check database health
- `Stats()`: Get pool statistics

**Features**:

- Connection pooling with configurable limits
- Health checks
- Graceful shutdown
- Error handling
- Statistics monitoring

#### `/pkg/logger`

**Purpose**: Application logging

**Contains**:

- `logger.go`: Logger implementation using Logrus
- `logger_test.go`: Unit tests

**Key Features**:

- Structured logging with fields
- Multiple log levels (debug, info, warn, error)
- JSON output support
- Formatted text output

#### `/pkg/validate`

**Purpose**: Data validation utilities

**Contains**: Validation helpers and utilities

### `/internal/bcrypt`

**Purpose**: Reusable internal secuiry

**Contains**:

#### `/internal/bcrypt/hash_password`

**Purpose**: Password hashing and verification security

**Contains**:

- `hash_password.go`: Argon2 password hashing implementation
- `hash_password_test.go`: Comprehensive unit and benchmark tests

**Key Functions**:

- `HashPassword(password string) (string, error)`: Hash password using Argon2-ID
- `CheckPassword(password, hashedPassword string) bool`: Verify password against hash

**Argon2 Configuration**:

- Algorithm: Argon2-ID (resistant to both GPU and side-channel attacks)
- Time cost: 3 iterations
- Memory cost: 64 MB
- Parallelism: 4 threads
- Tag length: 32 bytes (256-bit hash)
- Salt: Fixed for deterministic hashing

**Performance (Apple M5 Benchmark)**:

- HashPassword: ~27.5ms per operation
- CheckPassword: ~26.4ms per operation
- Suitable for typical web applications

**Testing**:

- Unit tests: Valid/invalid password handling
- Consistency tests: Verify deterministic hashing
- Benchmarks: Performance profiling included

### `/migrations`

**Purpose**: Database schema versioning

**Contains**:

- SQL migration files with timestamp prefixes
- `202602181953_User.sql`: User table schema

**Design**:

- Each file represents a version-controlled schema change
- Timestamp ensures ordered execution
- Can be applied incrementally

## Design Patterns Used

### 1. Layered Architecture

- **Separation of Concerns**: Each layer has distinct responsibility
- **Dependency Flow**: Downward dependency (upper layers depend on lower)
- **Testability**: Each layer can be tested independently

### 2. Dependency Injection

Database connection is injected into handlers:

```go
db, _ := config.NewDB(ctx, databaseURL)
// Pass db to handlers as needed
```

### 3. Repository Pattern (Data Access)

`DBPool` acts as repository, abstracting database access:

```go
// Simple interface for database operations
type DBPool struct {
    pool *pgxpool.Pool
}

func (db *DBPool) Query(ctx, sql, args...)
func (db *DBPool) Exec(ctx, sql, args...)
```

### 4. Middleware Pattern

Error handling middleware wraps all requests:

```go
// Global error handling
middleware.ErrorHandler(req, res)
```

### 5. Factory Pattern

Database initialization uses factory pattern:

```go
db, err := database.NewDB(ctx, databaseURL)
// Returns fully configured DBPool
```

## Configuration Management

### Environment-Based Configuration

```go
type Config struct {
    URL           string
    Port          int
    Env           string
    LogLevel      string
    DatabaseURL   string
    ReadTimeout   time.Duration
    WriteTimeout  time.Duration
}
```

### Loading Priority

1. `.env` file (highest priority)
2. Environment variables
3. Default values (lowest priority)

## Database Layer Design

### Connection Pool Architecture

```
┌──────────────────────────────────────┐
│        Application Code              │
│   (Handlers, Services, Queries)      │
└──────────────┬───────────────────────┘
               │
┌──────────────▼───────────────────────┐
│         DBPool Wrapper               │
│  (Query, Exec, Begin, Health)        │
└──────────────┬───────────────────────┘
               │
┌──────────────▼───────────────────────┐
│     pgxpool.Pool (5-25 connections)  │
│  (Connection Management & Pooling)   │
└──────────────┬───────────────────────┘
               │
┌──────────────▼───────────────────────┐
│       PostgreSQL Database            │
└──────────────────────────────────────┘
```

### Connection Lifecycle

1. **Opening**: Pool opens connections on demand (min 5)
2. **Reuse**: Connections are reused for subsequent queries
3. **Idling**: Idle connections kept for 2 minutes
4. **Closing**: Connections closed when idle or max lifetime exceeded
5. **Graceful Shutdown**: All connections closed on application exit

## Request/Response Flow

### Typical Request Flow

```
1. HTTP Request arrives
   ↓
2. Fiber Router matches route
   ↓
3. Middleware chain executes
   - Error handling setup
   - Logging setup
   ↓
4. Handler processes request
   - Validates input
   - Calls business logic
   ↓
5. Handler queries database (if needed)
   - Uses db.Query() or db.Exec()
   ↓
6. Handler formats response
   - Success: JSON data
   - Error: Error response with status code
   ↓
7. Middleware error handler (if error)
   - Formats error response
   - Sets status code
   ↓
8. HTTP Response sent to client
```

### Error Response Format

```json
{
  "code": 400,
  "message": "Bad Request",
  "status": "error"
}
```

## Testing Strategy

### Unit Tests

- **Location**: `*_test.go` files alongside source
- **Scope**: Test individual functions/methods
- **Dependencies**: Mock external dependencies

### Integration Tests

- **Location**: `*_test.go` with `// +build integration` tag
- **Scope**: Test interactions between components
- **Dependencies**: Use real PostgreSQL database (skipped if unavailable)

### Test Organization

```
pkg/
├── database/
│   ├── database.go
│   └── database_test.go      # Unit & Integration tests
├── logger/
│   ├── logger.go
│   └── logger_test.go        # Unit tests
```

## Scalability Considerations

### 1. Connection Pool Tuning

- Adjust `MaxConns` based on concurrent users
- Current: 25 connections (suitable for most applications)
- Monitor with `db.Stats()`

### 2. Database Performance

- Use indexes on frequently queried columns
- Avoid N+1 queries with proper joins
- Consider caching for read-heavy operations

### 3. Concurrent Request Handling

- Fiber automatically handles concurrent requests
- Connection pool manages database concurrency
- Each request gets its own context

### 4. Graceful Shutdown

- Connections are properly closed
- In-flight requests complete before shutdown
- Database changes are committed

## Security Considerations

### 1. SQL Injection Prevention

- Always use parameterized queries
- pgx prevents SQL injection through parameters

```go
// Safe: Uses parameterized query
db.Query(ctx, "SELECT * FROM users WHERE id = $1", id)

// Unsafe: String concatenation
db.Query(ctx, "SELECT * FROM users WHERE id = " + id)
```

### 2. Connection String Security

- Store `DATABASE_URL` in environment variables
- Never hardcode credentials
- Use `.env` file for local development (not in version control)

### 3. Error Information Leakage

- Error middleware sanitizes error messages
- Database errors not exposed to clients
- Logging includes full details for debugging

### 4. Password Security

- Passwords hashed using Argon2-ID (memory-hard algorithm)
- Never stored in plain text
- Hashed passwords never returned in API responses
- Separate password from user object in JSON marshaling

```go
type User struct {
    Password string `json:"-"` // Excluded from JSON output
}
```

- Validation enforces minimum 8 characters
- Consider implementing rate limiting for login attempts
- Use random salt per password (currently uses fixed salt for deterministic hashing)

### 5. Input Validation

- All user inputs validated before processing
- Using go-playground/validator for robust validation
- Field-level validation rules applied
- Detailed validation error messages for debugging
- Email format validation prevents invalid emails
- Username constraints prevent injection attacks

### Database Operations

| Operation          | Typical Latency | Notes                    |
| ------------------ | --------------- | ------------------------ |
| Single Row Query   | 1-5ms           | With index               |
| Range Query        | 5-50ms          | Depends on result size   |
| Insert/Update      | 2-10ms          | With constraints checked |
| Transaction Commit | 5-20ms          | Depends on complexity    |

### Connection Pool

| Metric           | Value          |
| ---------------- | -------------- |
| Min Connections  | 5              |
| Max Connections  | 25             |
| Connection Reuse | ~99% (typical) |
| Pool Acquisition | <1ms (cached)  |

## Authentication & Validation Architecture

### Signup Request Validation

Field validation using `go-playground/validator/v10`:

```go
type SignupRequest struct {
    Email    string `validate:"required,email"`              // Required, valid email
    Password string `validate:"required,min=8"`              // Required, min 8 chars
    FullName string `validate:"required,max=255"`            // Required, max 255 chars
    Username string `validate:"required,min=3,max=100"`      // Required, 3-100 chars
}
```

**Validation Error Response**:

```json
{
  "error": "Validation failed",
  "errors": [
    {
      "field": "Email",
      "message": "Email must be a valid email address"
    }
  ]
}
```

### Password Hashing Strategy

- **Algorithm**: Argon2-ID (OWASP recommended)
- **Library**: `golang.org/x/crypto/argon2`
- **Flow**:
  1. User submits plain text password
  2. Validate password meets requirements (min 8 chars)
  3. Hash with Argon2-ID (3 iterations, 64MB memory, 4 parallelism)
  4. Store hash in database
  5. Never store or transmit plain text password
  6. On login, hash submitted password and compare hashes

### Signup Service Architecture

```
SignupRequest
    ↓
[Validation] - go-playground/validator
    ├─ Email format
    ├─ Password length
    ├─ Username length
    └─ FullName length
    ↓
[Password Hashing] - Argon2-ID
    ├─ Hash password
    └─ Create User object with hash
    ↓
[Data Persistence] - Repository
    ├─ Save user to database
    ├─ Generate UUID
    ├─ Set timestamps
    └─ Return User object
    ↓
HTTP 201 Created with User Data
```

## Future Enhancement Opportunities

### 1. Password Security Enhancements

- Use random salt per password (currently uses fixed salt for deterministic hashing)
- Implement bcrypt/scrypt as alternative to Argon2
- Add password strength meter on frontend
- Implement password history to prevent reuse
- Add password expiration policies
- Implement secure password reset with email verification

### 2. Authentication Features

- JWT token generation for authenticated requests
- Refresh token mechanism
- Email verification for signup
- Password reset functionality via email
- Multi-factor authentication (MFA)
- Rate limiting on login/signup endpoints
- Account lockout after failed attempts
- CORS configuration for cross-origin requests

### 3. Caching Layer

- Add Redis for frequently accessed data
- Reduce database load
- Cache validated JWT tokens
- Improve response times

### 4. Query Builder

- Add query builder for complex queries
- Reduce repetitive SQL writing
- Improve maintainability

### 5. ORM Integration

- Consider GORM or sqlc for type-safe queries
- Reduce boilerplate code
- Better schema synchronization

### 6. Connection Monitoring

- Add metrics for pool health
- Create alerts for connection issues
- Dashboard for monitoring

### 7. Audit Logging

- Track data changes
- Maintain audit trail
- Compliance requirements
- Log signup attempts and authentication events
- Track failed login attempts

### 8. User Role & Permission System

- Define user roles (admin, user, moderator)
- Implement permission-based access control
- Role-based route protection

## Testing Infrastructure

### Hash Password Tests

Location: `/internal/security/hash_password/hash_password_test.go`

**Test Coverage**:

- Valid password hashing with various lengths
- Empty password error handling
- Unicode character support
- Consistency verification (deterministic hashing)
- Password verification matching
- Case sensitivity
- Benchmark tests for performance tracking

**Running Tests**:

```bash
# Run all tests
go test ./internal/security/hash_password -v

# Run benchmarks
go test ./internal/security/hash_password -bench=. -benchmem

# Run specific test
go test ./internal/security/hash_password -run TestHashPassword -v
```

## References

- [Fiber Web Framework](https://gofiber.io/)
- [pgx PostgreSQL Driver](https://github.com/jackc/pgx)
- [Logrus Logger](https://github.com/sirupsen/logrus)
- [Go Crypto - Argon2](https://pkg.go.dev/golang.org/x/crypto/argon2)
- [go-playground/validator](https://github.com/go-playground/validator)
- [Go Best Practices](https://golang.org/doc/effective_go)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [OWASP Input Validation Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Input_Validation_Cheat_Sheet.html)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
