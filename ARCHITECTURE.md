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
- `common/`: Common utilities and handlers
  - `router.go`: Request routing configuration
  - `health/`: Health check endpoints
  - `home/`: Home/welcome endpoints
- `examples/`: Example endpoints and routes

**Design Pattern**: Domain-Driven Design
- Each domain has clear responsibility
- Routes registered in domain modules
- Handlers process specific domain requests

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

## Performance Characteristics

### Database Operations
| Operation | Typical Latency | Notes |
|-----------|-----------------|-------|
| Single Row Query | 1-5ms | With index |
| Range Query | 5-50ms | Depends on result size |
| Insert/Update | 2-10ms | With constraints checked |
| Transaction Commit | 5-20ms | Depends on complexity |

### Connection Pool
| Metric | Value |
|--------|-------|
| Min Connections | 5 |
| Max Connections | 25 |
| Connection Reuse | ~99% (typical) |
| Pool Acquisition | <1ms (cached) |

## Future Enhancement Opportunities

### 1. Caching Layer
- Add Redis for frequently accessed data
- Reduce database load
- Improve response times

### 2. Query Builder
- Add query builder for complex queries
- Reduce repetitive SQL writing
- Improve maintainability

### 3. ORM Integration
- Consider GORM or sqlc for type-safe queries
- Reduce boilerplate code
- Better schema synchronization

### 4. Connection Monitoring
- Add metrics for pool health
- Create alerts for connection issues
- Dashboard for monitoring

### 5. Audit Logging
- Track data changes
- Maintain audit trail
- Compliance requirements

## References

- [Fiber Web Framework](https://gofiber.io/)
- [pgx PostgreSQL Driver](https://github.com/jackc/pgx)
- [Logrus Logger](https://github.com/sirupsen/logrus)
- [Go Best Practices](https://golang.org/doc/effective_go)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
