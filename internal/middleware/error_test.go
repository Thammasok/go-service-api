package middleware

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStatusMessage tests the statusMessage helper function.
func TestStatusMessage(t *testing.T) {
	t.Run("status messages", func(t *testing.T) {
		tests := []struct {
			code     int
			expected string
		}{
			{fiber.StatusBadRequest, "bad_request"},
			{fiber.StatusUnauthorized, "unauthorized"},
			{fiber.StatusForbidden, "forbidden"},
			{fiber.StatusNotFound, "not_found"},
			{fiber.StatusInternalServerError, "internal_error"},
			{fiber.StatusServiceUnavailable, "service_unavailable"},
			{200, "error"},
		}

		for _, tt := range tests {
			t.Run("code_"+string(rune(tt.code)), func(t *testing.T) {
				got := statusMessage(tt.code)
				assert.Equal(t, tt.expected, got)
			})
		}
	})
}

// TestErrorResponseJSON tests that ErrorResponse marshals correctly to JSON.
func TestErrorResponseJSON(t *testing.T) {
	t.Run("error response JSON", func(t *testing.T) {
		resp := ErrorResponse{
			Error:   "test_error",
			Message: "Test message",
			Code:    400,
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err, "marshal should not error")

		var decoded ErrorResponse
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err, "unmarshal should not error")

		assert.Equal(t, resp, decoded, "unmarshaled response should match original")
	})
}

// TestValidationErrorResponse tests the validation error helper.
func TestValidationErrorResponse(t *testing.T) {
	t.Run("validation error response", func(t *testing.T) {
		app := fiber.New()

		app.Post("/test", func(c fiber.Ctx) error {
			return ValidationErrorResponse(c, "name is required")
		})

		req, _ := http.NewRequest("POST", "/test", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var respBody ErrorResponse
		err := json.NewDecoder(resp.Body).Decode(&respBody)
		require.NoError(t, err, "decode should not error")

		assert.Equal(t, "validation_error", respBody.Error)
		assert.Equal(t, "name is required", respBody.Message)
	})
}

// TestAuthErrorResponse tests the auth error helper.
func TestAuthErrorResponse(t *testing.T) {
	t.Run("auth error response", func(t *testing.T) {
		app := fiber.New()

		app.Get("/test", func(c fiber.Ctx) error {
			return AuthErrorResponse(c, "invalid token")
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		var respBody ErrorResponse
		err := json.NewDecoder(resp.Body).Decode(&respBody)
		require.NoError(t, err, "decode should not error")

		assert.Equal(t, "unauthorized", respBody.Error)
	})
}

// TestNotFoundResponse tests the not found error helper.
func TestNotFoundResponse(t *testing.T) {
	t.Run("not found response", func(t *testing.T) {
		app := fiber.New()

		app.Get("/test", func(c fiber.Ctx) error {
			return NotFoundResponse(c, "user not found")
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		var respBody ErrorResponse
		err := json.NewDecoder(resp.Body).Decode(&respBody)
		require.NoError(t, err, "decode should not error")

		assert.Equal(t, "not_found", respBody.Error)
	})
}

// TestInternalErrorResponse tests the internal error helper.
func TestInternalErrorResponse(t *testing.T) {
	t.Run("internal error response", func(t *testing.T) {
		app := fiber.New()

		app.Get("/test", func(c fiber.Ctx) error {
			return InternalErrorResponse(c, "database connection failed")
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		var respBody ErrorResponse
		err := json.NewDecoder(resp.Body).Decode(&respBody)
		require.NoError(t, err, "decode should not error")

		assert.Equal(t, "internal_error", respBody.Error)
	})
}

// BenchmarkStatusMessage benchmarks the statusMessage function.
func BenchmarkStatusMessage(b *testing.B) {
	b.Run("benchmarks message lookup", func(b *testing.B) {
		codes := []int{
			fiber.StatusBadRequest,
			fiber.StatusUnauthorized,
			fiber.StatusNotFound,
			fiber.StatusInternalServerError,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, code := range codes {
				_ = statusMessage(code)
			}
		}
	})
}

// BenchmarkErrorResponseMarshal benchmarks JSON marshaling of ErrorResponse.
func BenchmarkErrorResponseMarshal(b *testing.B) {
	b.Run("error response marshal", func(b *testing.B) {
		resp := ErrorResponse{
			Error:   "test_error",
			Message: "Error message",
			Code:    500,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(resp)
		}
	})
}
