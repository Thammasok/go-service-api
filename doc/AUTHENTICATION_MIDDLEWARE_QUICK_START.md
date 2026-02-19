# AuthMiddleware Quick Start Guide

## What was created?

### 1. Auth Middleware (`internal/middleware/auth_middleware.go`)
- `AuthMiddleware(tm *token.TokenManager)`: Main middleware function
- `GetUserIDFromContext(c fiber.Ctx)`: Helper to retrieve user ID from context
- `extractBearerToken(authHeader string)`: Internal function for token parsing

### 2. Refresh Token Handler (`internal/domain/authentication/refresh_token_handler.go`)
- `RefreshTokenHandler(db, cfg)`: Handles token refresh requests
- Validates refresh tokens and generates new access tokens

### 3. Protected Routes Example (`internal/domain/private/`)
- `private_route.go`: Route setup with authentication middleware
- `profile_handler.go`: Example handler showing how to use authenticated context

## Quick Start

### Step 1: Setup Protected Routes

```go
// In your route file (e.g., private_route.go)
import (
    "dvith.com/go-service-api/internal/middleware"
    "dvith.com/go-service-api/internal/security/token"
)

func Routers(app fiber.Router, db *database.DBPool, cfg config.Config) {
    tm := token.NewTokenManager(token.TokenConfig{
        SecretKey:       cfg.JWTSecretKey,
        ExpirationTime:  cfg.JWTExpirationTime,
        RefreshDuration: cfg.JWTRefreshDuration,
        Issuer:          cfg.JWTIssuer,
    })

    // Routes requiring authentication
    withAuth := app.Group("/protected", middleware.AuthMiddleware(tm))
    withAuth.Get("/profile", ProfileHandler(db))
}
```

### Step 2: Access User ID in Handlers

```go
func MyHandler(db *database.DBPool) fiber.Handler {
    return func(c fiber.Ctx) error {
        // Get authenticated user ID
        userID, err := middleware.GetUserIDFromContext(c)
        if err != nil {
            return middleware.AuthErrorResponse(c, "user not authenticated")
        }

        // Use userID (e.g., query database)
        user, _ := db.GetUser(context.Background(), userID)
        
        return c.JSON(user)
    }
}
```

### Step 3: Client Makes Request

```bash
# 1. Login first to get tokens
curl -X POST http://localhost:8080/api/v1/signin \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com", "password":"password123"}'

# Response includes: access_token, refresh_token

# 2. Use access_token in Authorization header
curl -X GET http://localhost:8080/api/v1/protected/profile \
  -H "Authorization: Bearer <access_token>"

# 3. When access_token expires, use refresh_token
curl -X POST http://localhost:8080/api/v1/refresh-token \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

## API Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/signup` | Register new user | No |
| POST | `/signin` | Login user | No |
| POST | `/refresh-token` | Get new access token | No |
| GET | `/protected/profile` | Get user profile | Yes |

## Environment Variables

```env
# JWT Configuration
JWT_SECRET_KEY=your-secret-key-change-in-production
JWT_EXPIRATION_TIME=1h
JWT_REFRESH_DURATION=168h
JWT_ISSUER=go-service-api
```

## Error Handling

The middleware returns 401 Unauthorized for:
- Missing Authorization header
- Invalid Bearer token format
- Invalid or expired access token

Example error response:
```json
{
  "error": "unauthorized",
  "message": "invalid or expired access token",
  "code": 401
}
```

## Key Features

✅ JWT access token validation  
✅ Bearer token extraction  
✅ User context storage  
✅ Automatic token expiration checks  
✅ Refresh token support  
✅ Consistent error responses  
✅ Structured logging  

## Testing with Bruno

1. Install Bruno API client
2. Open `doc/api-doc/bruno.json`
3. Create new request or test existing endpoints:
   - Use `/signup` or `/signin` to get tokens
   - Copy `access_token` from response
   - Set header: `Authorization: Bearer {access_token}`
   - Test protected endpoints

## Next Steps

1. Extend protected routes with your endpoints
2. Configure JWT secret for production
3. Implement token refresh logic in frontend
4. Add roles-based access control if needed
5. Consider token blacklisting for logout
