package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dvith.com/go-service-api/internal/security/token"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test fixtures
func createTestTokenManager() *token.TokenManager {
	return token.NewTokenManager(token.TokenConfig{
		SecretKey:       "test-secret-key-for-testing",
		ExpirationTime:  1 * time.Hour,
		RefreshDuration: 7 * 24 * time.Hour,
		Issuer:          "go-service-api",
	})
}

// TestExtractBearerToken tests the extractBearerToken helper function
func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    string
		wantErr bool
	}{
		{
			name:    "valid bearer token",
			header:  "Bearer valid-token-string",
			want:    "valid-token-string",
			wantErr: false,
		},
		{
			name:    "missing bearer prefix",
			header:  "Basic dXNlcm5hbWU6cGFzc3dvcmQ=",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty token",
			header:  "Bearer ",
			want:    "",
			wantErr: true,
		},
		{
			name:    "no space in header",
			header:  "Bearertoken",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty header",
			header:  "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "only bearer",
			header:  "Bearer",
			want:    "",
			wantErr: true,
		},
		{
			name:    "multiple spaces",
			header:  "Bearer token with spaces",
			want:    "token with spaces",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractBearerToken(tt.header)
			if tt.wantErr {
				assert.Error(t, err, "extractBearerToken should return error")
			} else {
				require.NoError(t, err, "extractBearerToken should not error")
				assert.Equal(t, tt.want, got, "extracted token mismatch")
			}
		})
	}
}

// TestAuthMiddleware_ValidToken tests middleware with valid token
func TestAuthMiddleware_ValidToken(t *testing.T) {
	tm := createTestTokenManager()
	userID := uuid.New()

	// Generate valid token
	accessToken, err := tm.GenerateAccessToken(userID)
	require.NoError(t, err, "failed to generate access token")

	// Create test app and middleware
	app := fiber.New()
	app.Use(AuthMiddleware(tm))

	// Test route
	app.Get("/protected", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Create request with valid token
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	// Execute request
	resp, err := app.Test(req)
	require.NoError(t, err, "test request failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected 200 OK status")
}

// TestAuthMiddleware_MissingAuthHeader tests middleware without authorization header
func TestAuthMiddleware_MissingAuthHeader(t *testing.T) {
	tm := createTestTokenManager()

	app := fiber.New()
	app.Use(ErrorHandler(), AuthMiddleware(tm))

	app.Get("/protected", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Create request without authorization header
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)

	resp, err := app.Test(req)
	require.NoError(t, err, "test request failed")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expected 401 Unauthorized status")
}

// TestAuthMiddleware_InvalidBearerFormat tests middleware with invalid bearer token format
func TestAuthMiddleware_InvalidBearerFormat(t *testing.T) {
	tm := createTestTokenManager()

	app := fiber.New()
	app.Use(ErrorHandler(), AuthMiddleware(tm))

	app.Get("/protected", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	tests := []struct {
		name      string
		authValue string
	}{
		{"basic auth", "Basic dXNlcm5hbWU6cGFzc3dvcmQ="},
		{"missing bearer word", "dXNlcm5hbWU6cGFzc3dvcmQ="},
		{"empty bearer", "Bearer "},
		{"only bearer", "Bearer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", tt.authValue)

			resp, err := app.Test(req)
			require.NoError(t, err, "test request failed")
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expected 401 Unauthorized")
		})
	}
}

// TestAuthMiddleware_InvalidToken tests middleware with invalid token
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	tm := createTestTokenManager()

	app := fiber.New()
	app.Use(ErrorHandler(), AuthMiddleware(tm))

	app.Get("/protected", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Create request with invalid token
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.string")

	resp, err := app.Test(req)
	require.NoError(t, err, "test request failed")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expected 401 Unauthorized")
}

// TestAuthMiddleware_ExpiredToken tests middleware with expired token
func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	// Create token manager with very short expiration
	tm := token.NewTokenManager(token.TokenConfig{
		SecretKey:       "test-secret-key",
		ExpirationTime:  1 * time.Millisecond,
		RefreshDuration: 7 * 24 * time.Hour,
		Issuer:          "go-service-api",
	})

	userID := uuid.New()
	expiredToken, err := tm.GenerateAccessToken(userID)
	require.NoError(t, err, "failed to generate token")

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	app := fiber.New()
	app.Use(ErrorHandler(), AuthMiddleware(tm))

	app.Get("/protected", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", expiredToken))

	resp, err := app.Test(req)
	require.NoError(t, err, "test request failed")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expected 401 for expired token")
}

// TestAuthMiddleware_TokenFromWrongKey tests token signed with different key
func TestAuthMiddleware_TokenFromWrongKey(t *testing.T) {
	// Create token with one key
	tm1 := token.NewTokenManager(token.TokenConfig{
		SecretKey:       "first-secret-key",
		ExpirationTime:  1 * time.Hour,
		RefreshDuration: 7 * 24 * time.Hour,
		Issuer:          "go-service-api",
	})

	userID := uuid.New()
	accessToken, err := tm1.GenerateAccessToken(userID)
	require.NoError(t, err, "failed to generate token")

	// Try to validate with different key
	tm2 := token.NewTokenManager(token.TokenConfig{
		SecretKey:       "different-secret-key",
		ExpirationTime:  1 * time.Hour,
		RefreshDuration: 7 * 24 * time.Hour,
		Issuer:          "go-service-api",
	})

	app := fiber.New()
	app.Use(ErrorHandler(), AuthMiddleware(tm2))

	app.Get("/protected", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := app.Test(req)
	require.NoError(t, err, "test request failed")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expected 401 for token signed with wrong key")
}

// TestAuthMiddleware_ContextStorage tests that user ID is stored in context
func TestAuthMiddleware_ContextStorage(t *testing.T) {
	tm := createTestTokenManager()
	userID := uuid.New()

	accessToken, err := tm.GenerateAccessToken(userID)
	require.NoError(t, err, "failed to generate token")

	app := fiber.New()
	app.Use(AuthMiddleware(tm))

	// Handler that retrieves user ID from context
	app.Get("/protected", func(c fiber.Ctx) error {
		retrievedUserID, err := GetUserIDFromContext(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Return the user ID to verify it matches
		return c.JSON(fiber.Map{
			"user_id": retrievedUserID.String(),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := app.Test(req)
	require.NoError(t, err, "test request failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected 200 OK")
}

// TestGetUserIDFromContext tests the GetUserIDFromContext helper function
func TestGetUserIDFromContext(t *testing.T) {
	app := fiber.New()

	t.Run("valid user id in context", func(t *testing.T) {
		userID := uuid.New()

		app.Get("/test", func(c fiber.Ctx) error {
			c.Locals(ContextKeyUserID, userID)

			retrieved, err := GetUserIDFromContext(c)
			require.NoError(t, err, "GetUserIDFromContext should not error")
			assert.Equal(t, userID, retrieved, "retrieved user ID should match")

			return c.JSON(fiber.Map{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("missing user id in context", func(t *testing.T) {
		app2 := fiber.New()
		app2.Get("/test", func(c fiber.Ctx) error {
			_, err := GetUserIDFromContext(c)
			assert.Error(t, err, "GetUserIDFromContext should error when user_id not in context")
			assert.Contains(t, err.Error(), "not found in context")

			return c.JSON(fiber.Map{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp, err := app2.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("invalid user id type in context", func(t *testing.T) {
		app3 := fiber.New()
		app3.Get("/test", func(c fiber.Ctx) error {
			// Store wrong type
			c.Locals(ContextKeyUserID, "not-a-uuid")

			_, err := GetUserIDFromContext(c)
			assert.Error(t, err, "GetUserIDFromContext should error with wrong type")
			assert.Contains(t, err.Error(), "invalid user_id type")

			return c.JSON(fiber.Map{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp, err := app3.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

// TestAuthMiddleware_MultipleRequests tests middleware with multiple sequential requests
func TestAuthMiddleware_MultipleRequests(t *testing.T) {
	tm := createTestTokenManager()

	app := fiber.New()
	app.Use(AuthMiddleware(tm))

	app.Get("/protected", func(c fiber.Ctx) error {
		userID, err := GetUserIDFromContext(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"user_id": userID.String()})
	})

	// Test with different users
	for i := 0; i < 3; i++ {
		userID := uuid.New()
		accessToken, err := tm.GenerateAccessToken(userID)
		require.NoError(t, err, "failed to generate token")

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

		resp, err := app.Test(req)
		require.NoError(t, err, "test request failed")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "request %d should succeed", i+1)
	}
}

// TestAuthMiddleware_DifferentHTTPMethods tests middleware with various HTTP methods
func TestAuthMiddleware_DifferentHTTPMethods(t *testing.T) {
	tm := createTestTokenManager()
	userID := uuid.New()

	accessToken, err := tm.GenerateAccessToken(userID)
	require.NoError(t, err, "failed to generate token")

	app := fiber.New()
	app.Use(AuthMiddleware(tm))

	// Register handlers for different methods
	app.Get("/protected", func(c fiber.Ctx) error { return c.JSON(fiber.Map{"method": "GET"}) })
	app.Post("/protected", func(c fiber.Ctx) error { return c.JSON(fiber.Map{"method": "POST"}) })
	app.Put("/protected", func(c fiber.Ctx) error { return c.JSON(fiber.Map{"method": "PUT"}) })
	app.Delete("/protected", func(c fiber.Ctx) error { return c.JSON(fiber.Map{"method": "DELETE"}) })

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/protected", nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode, "request with %s should succeed", method)
		})
	}
}

// BenchmarkAuthMiddleware benchmarks the middleware performance
func BenchmarkAuthMiddleware(b *testing.B) {
	tm := createTestTokenManager()
	userID := uuid.New()

	accessToken, err := tm.GenerateAccessToken(userID)
	if err != nil {
		b.Fatalf("failed to generate token: %v", err)
	}

	app := fiber.New()
	app.Use(AuthMiddleware(tm))

	app.Get("/protected", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		app.Test(req)
	}
}
