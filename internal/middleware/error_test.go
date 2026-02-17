package middleware

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// TestStatusMessage tests the statusMessage helper function.
func TestStatusMessage(t *testing.T) {
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
			if got != tt.expected {
				t.Errorf("statusMessage(%d) = %q, want %q", tt.code, got, tt.expected)
			}
		})
	}
}

// TestErrorResponseJSON tests that ErrorResponse marshals correctly to JSON.
func TestErrorResponseJSON(t *testing.T) {
	resp := ErrorResponse{
		Error:   "test_error",
		Message: "Test message",
		Code:    400,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal ErrorResponse: %v", err)
	}

	var decoded ErrorResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ErrorResponse: %v", err)
	}

	if decoded.Error != resp.Error || decoded.Message != resp.Message || decoded.Code != resp.Code {
		t.Errorf("unmarshal mismatch: got %+v, want %+v", decoded, resp)
	}
}

// TestValidationErrorResponse tests the validation error helper.
func TestValidationErrorResponse(t *testing.T) {
	app := fiber.New()

	app.Post("/test", func(c fiber.Ctx) error {
		return ValidationErrorResponse(c, "name is required")
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status %d, got %d", fiber.StatusBadRequest, resp.StatusCode)
	}

	var respBody ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if respBody.Error != "validation_error" {
		t.Errorf("expected error 'validation_error', got %q", respBody.Error)
	}
	if respBody.Message != "name is required" {
		t.Errorf("expected message 'name is required', got %q", respBody.Message)
	}
}

// TestAuthErrorResponse tests the auth error helper.
func TestAuthErrorResponse(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c fiber.Ctx) error {
		return AuthErrorResponse(c, "invalid token")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}

	var respBody ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if respBody.Error != "unauthorized" {
		t.Errorf("expected error 'unauthorized', got %q", respBody.Error)
	}
}

// TestNotFoundResponse tests the not found error helper.
func TestNotFoundResponse(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c fiber.Ctx) error {
		return NotFoundResponse(c, "user not found")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("expected status %d, got %d", fiber.StatusNotFound, resp.StatusCode)
	}

	var respBody ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if respBody.Error != "not_found" {
		t.Errorf("expected error 'not_found', got %q", respBody.Error)
	}
}

// TestInternalErrorResponse tests the internal error helper.
func TestInternalErrorResponse(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c fiber.Ctx) error {
		return InternalErrorResponse(c, "database connection failed")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", fiber.StatusInternalServerError, resp.StatusCode)
	}

	var respBody ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if respBody.Error != "internal_error" {
		t.Errorf("expected error 'internal_error', got %q", respBody.Error)
	}
}

// BenchmarkStatusMessage benchmarks the statusMessage function.
func BenchmarkStatusMessage(b *testing.B) {
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
}

// BenchmarkErrorResponseMarshal benchmarks JSON marshaling of ErrorResponse.
func BenchmarkErrorResponseMarshal(b *testing.B) {
	resp := ErrorResponse{
		Error:   "test_error",
		Message: "Error message",
		Code:    500,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(resp)
	}
}
