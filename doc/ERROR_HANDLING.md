# Error Handling Guide

## Overview

The error handling middleware (`internal/middleware/error_middleware.go`) provides centralized error handling for all `/api/v1` routes. It automatically catches panics, logs errors, and returns consistent JSON error responses.

## How It Works

1. **Automatic Panic Recovery**: Any panics in handlers are caught and converted to 500 responses
2. **Consistent Error Format**: All errors return a uniform `ErrorResponse` JSON structure with `error`, `message`, and `code` fields
3. **Logging**: All errors are automatically logged via the `pkg` logger
4. **Transparent**: The middleware is already applied to all `/api/v1` routes via `internal/domain/mod.go`

## ErrorResponse Structure

```json
{
  "error": "error_type",
  "message": "human readable message",
  "code": 400
}
```

## Usage Examples

### 1. Validation Error (400 Bad Request)

```go
func CreateUserHandler(c fiber.Ctx) error {
    var req CreateUserRequest
    
    if err := c.BindJSON(&req); err != nil {
        // Return 400 with validation_error type
        return middleware.ValidationErrorResponse(c, "invalid request: "+err.Error())
    }
    
    if req.Name == "" {
        return middleware.ValidationErrorResponse(c, "name is required")
    }
    
    // Success
    return c.JSON(fiber.Map{"id": 1, "name": req.Name})
}
```

**HTTP Response:**
```
Status: 400
Body: {
  "error": "validation_error",
  "message": "name is required",
  "code": 400
}
```

### 2. Authorization Error (401 Unauthorized)

```go
func ProtectedHandler(c fiber.Ctx) error {
    token := c.Get("Authorization")
    
    if !isValidToken(token) {
        // Return 401 with unauthorized error
        return middleware.AuthErrorResponse(c, "invalid token")
    }
    
    // Success
    return c.JSON(fiber.Map{"data": "secret"})
}
```

**HTTP Response:**
```
Status: 401
Body: {
  "error": "unauthorized",
  "message": "invalid token",
  "code": 401
}
```

### 3. Not Found Error (404)

```go
func GetUserHandler(c fiber.Ctx) error {
    userID := c.Params("id")
    user := findUser(userID) // returns nil if not found
    
    if user == nil {
        // Return 404 with not_found error
        return middleware.NotFoundResponse(c, "user with id "+userID+" not found")
    }
    
    return c.JSON(user)
}
```

**HTTP Response:**
```
Status: 404
Body: {
  "error": "not_found",
  "message": "user with id 999 not found",
  "code": 404
}
```

### 4. Internal Server Error (500)

```go
func FetchDataHandler(c fiber.Ctx) error {
    data, err := fetchFromDatabase()
    
    if err != nil {
        // Return 500 with internal_error (don't expose internal details)
        return middleware.InternalErrorResponse(c, "failed to retrieve data")
    }
    
    return c.JSON(data)
}
```

**HTTP Response:**
```
Status: 500
Body: {
  "error": "internal_error",
  "message": "failed to retrieve data",
  "code": 500
}
```

### 5. Panic Recovery (Automatic)

```go
func PanicHandler(c fiber.Ctx) error {
    // This panic is caught by ErrorHandler middleware automatically
    panic("something unexpected happened")
    
    // Will return:
    // Status: 500
    // Body: {
    //   "error": "internal_error",
    //   "message": "An unexpected error occurred",
    //   "code": 500
    // }
}
```

## Available Helper Functions

| Function | HTTP Status | Error Type | Use Case |
|----------|-------------|-----------|----------|
| `ValidationErrorResponse(c, msg)` | 400 | validation_error | Invalid input data |
| `AuthErrorResponse(c, msg)` | 401 | unauthorized | Authentication/token issues |
| `NotFoundResponse(c, msg)` | 404 | not_found | Resource doesn't exist |
| `InternalErrorResponse(c, msg)` | 500 | internal_error | Server errors |

## How Errors Get Logged

All errors are automatically logged with context via the `pkg` logger:

```
[2026-02-17T10:30:45Z] [ERROR] request error path=/api/v1/users method=POST code=400 error="name is required"
```

## Key Points to Remember

1. **Don't expose internal details** - Use `InternalErrorResponse` instead of returning raw error messages to clients
2. **Message field is optional** - If you don't include a message, it won't appear in the JSON response
3. **Panics are caught** - You don't need to handle panics explicitly; the middleware catches them
4. **Use appropriate status codes** - Pick the helper that matches the error type
5. **Logging is automatic** - All errors are logged, so you can debug using logs

## Testing Error Responses

You can test error responses directly by accessing the home endpoint with query parameters:

```bash
# Success
curl http://localhost:8080/api/v1/

# Validation error
curl "http://localhost:8080/api/v1/?mode=validation"

# Auth error
curl "http://localhost:8080/api/v1/?mode=auth"

# Not found error
curl "http://localhost:8080/api/v1/?mode=notfound"

# Internal error
curl "http://localhost:8080/api/v1/?mode=error"

# Panic (caught by middleware)
curl "http://localhost:8080/api/v1/?mode=panic"
```
