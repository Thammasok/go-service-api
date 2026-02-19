# Authentication Middleware Documentation

## Overview

The `AuthMiddleware` provides JWT-based access token validation for protecting routes in the API. It validates Bearer tokens from the Authorization header and stores the authenticated user's ID in the request context.

## Features

- **Access Token Validation**: Validates JWT access tokens using HS256 signing method
- **Bearer Token Extraction**: Automatically extracts tokens from `Authorization: Bearer <token>` header
- **User Context Storage**: Stores the authenticated user ID in route context for handler access
- **Error Handling**: Returns appropriate error responses for missing or invalid tokens
- **Refresh Token Support**: Provides endpoint to refresh expired access tokens

## Architecture

### Components

1. **AuthMiddleware**: Main middleware function that validates access tokens
2. **GetUserIDFromContext**: Helper function to retrieve user ID from context
3. **extractBearerToken**: Internal function to parse Bearer token format
4. **RefreshTokenHandler**: Endpoint to generate new access tokens using refresh tokens

### Token Flow

```
User Login (POST /signin or /signup)
    ↓
Generate Access Token + Refresh Token
    ↓
Client stores both tokens
    ↓
Client sends Access Token in Authorization header
    ↓
AuthMiddleware validates Access Token
    ↓
User ID stored in context
    ↓
Handler processes request
    ↓
(When Access Token expires)
    ↓
Client sends Refresh Token to /refresh-token
    ↓
Server validates Refresh Token
    ↓
New Access Token generated
    ↓
Client updates Access Token
```

## Usage

### 1. Protecting a Route with AuthMiddleware

In your route handler file (e.g., `private_route.go`):

```go
package private

import (
    "dvith.com/go-service-api/internal/config"
    "dvith.com/go-service-api/internal/middleware"
    "dvith.com/go-service-api/internal/security/token"
    "dvith.com/go-service-api/pkg/database"
    "github.com/gofiber/fiber/v3"
)

func Routers(app fiber.Router, db *database.DBPool, cfg config.Config) {
    // Initialize token manager
    tm := token.NewTokenManager(token.TokenConfig{
        SecretKey:       cfg.JWTSecretKey,
        ExpirationTime:  cfg.JWTExpirationTime,
        RefreshDuration: cfg.JWTRefreshDuration,
        Issuer:          cfg.JWTIssuer,
    })

    // Create protected route group
    withAuth := app.Group("/protected", middleware.AuthMiddleware(tm))

    // Add protected routes
    withAuth.Get("/profile", ProfileHandler(db))
    withAuth.Post("/settings", UpdateSettingsHandler(db))
}
```

### 2. Using User ID in Handlers

In your handler:

```go
func ProfileHandler(db *database.DBPool) fiber.Handler {
    return func(c fiber.Ctx) error {
        // Get user ID from context
        userID, err := middleware.GetUserIDFromContext(c)
        if err != nil {
            return middleware.AuthErrorResponse(c, "user not authenticated")
        }

        // Use userID to query database
        user, err := db.GetUser(context.Background(), userID)
        if err != nil {
            return middleware.InternalErrorResponse(c, "failed to fetch user")
        }

        return c.JSON(user)
    }
}
```

### 3. Client-Side: Sending Access Token

When making a request to a protected route:

```bash
curl -X GET http://localhost:8080/api/v1/protected/profile \
  -H "Authorization: Bearer <access_token>"
```

### 4. Refreshing Expired Access Token

**Endpoint**: `POST /api/v1/refresh-token`

**Request**:
```json
{
  "refresh_token": "<your_refresh_token>"
}
```

**Response**:
```json
{
  "access_token": "<new_access_token>",
  "refresh_token": "<refresh_token>",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

## Error Responses

### Missing Authorization Header
```json
{
  "error": "unauthorized",
  "message": "missing authorization header",
  "code": 401
}
```

### Invalid Token Format
```json
{
  "error": "unauthorized",
  "message": "invalid authorization header format",
  "code": 401
}
```

### Invalid or Expired Token
```json
{
  "error": "unauthorized",
  "message": "invalid or expired access token",
  "code": 401
}
```

## Configuration

The middleware uses JWT configuration from environment variables:

```env
# JWT Secret Key (change in production!)
JWT_SECRET_KEY=your-secret-key-change-in-production

# Access Token Expiration (default: 1 hour)
JWT_EXPIRATION_TIME=1h

# Refresh Token Expiration (default: 168 hours = 7 days)
JWT_REFRESH_DURATION=168h

# JWT Issuer
JWT_ISSUER=go-service-api
```

## Token Content

### Access Token Claims
```json
{
  "user_id": "uuid-here",
  "exp": 1613692800,
  "iat": 1613689200,
  "nbf": 1613689200,
  "iss": "go-service-api",
  "aud": ["go-service-api-users"]
}
```

### Refresh Token Claims
```json
{
  "user_id": "uuid-here",
  "exp": 1614297600,
  "iat": 1613689200,
  "nbf": 1613689200,
  "iss": "go-service-api",
  "aud": ["go-service-api-refresh"]
}
```

## Best Practices

1. **Secret Key Management**: Always use a strong random secret key in production
2. **Token Expiration**: Set short expiration for access tokens (15 minutes to 1 hour)
3. **Refresh Token Expiration**: Set longer expiration for refresh tokens (7-30 days)
4. **HTTPS Only**: Always use HTTPS in production to prevent token interception
5. **Token Storage**: Never store tokens in localStorage, use secure cookies instead
6. **Token Rotation**: Consider rotating refresh tokens after use for enhanced security
7. **Context Cleanup**: The middleware automatically cleans up context after request

## Testing

Example test request using Bruno API client:

1. Signin to get tokens:
```
POST /api/v1/signin
{
  "email": "user@example.com",
  "password": "password123"
}
```

2. Copy `access_token` from response

3. Use token in protected route:
```
GET /api/v1/protected/profile
Authorization: Bearer <access_token>
```

4. Refresh token when expired:
```
POST /api/v1/refresh-token
{
  "refresh_token": "<refresh_token>"
}
```

## Troubleshooting

### "missing authorization header"
- Ensure you're sending the `Authorization` header
- Format should be: `Authorization: Bearer <token>`

### "invalid authorization header format"
- Check if the header starts with `Bearer ` (with a space)
- Ensure there's exactly one space between `Bearer` and the token

### "invalid or expired access token"
- Token may have expired - use refresh token to get a new one
- Token may have been tampered with - regenerate by signing in
- Secret key may have changed - ensure same key is used for validation

### Context errors
- Ensure the route is protected with `middleware.AuthMiddleware(tm)`
- Check that you're calling `GetUserIDFromContext()` instead of accessing context directly
