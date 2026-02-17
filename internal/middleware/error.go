package middleware

import (
	"dvith.com/go-service-api/pkg/logger"
	"github.com/gofiber/fiber/v3"
)

// ErrorResponse is a uniform error response structure for the API.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

// ErrorHandler is middleware that catches panics and errors from route handlers,
// logs them, and returns a consistent JSON error response.
func ErrorHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Catch any panic from the handler
		defer func() {
			if r := recover(); r != nil {
				logger.Error("handler panic", map[string]any{
					"path":   c.Path(),
					"method": c.Method(),
					"panic":  r,
				})
				c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
					Error:   "internal_error",
					Message: "An unexpected error occurred",
					Code:    fiber.StatusInternalServerError,
				})
			}
		}()

		err := c.Next()

		// Handle Fiber errors
		if err != nil {
			var code int
			var errStr string

			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				errStr = e.Error()
			} else {
				code = fiber.StatusInternalServerError
				errStr = err.Error()
			}

			logger.Error("request error", map[string]any{
				"path":   c.Path(),
				"method": c.Method(),
				"code":   code,
				"error":  errStr,
			})

			// Get a simple status message
			statusMsg := statusMessage(code)
			return c.Status(code).JSON(ErrorResponse{
				Error:   statusMsg,
				Message: errStr,
				Code:    code,
			})
		}

		return nil
	}
}

// statusMessage returns a simple message for a given HTTP status code.
func statusMessage(code int) string {
	switch code {
	case fiber.StatusBadRequest:
		return "bad_request"
	case fiber.StatusUnauthorized:
		return "unauthorized"
	case fiber.StatusForbidden:
		return "forbidden"
	case fiber.StatusNotFound:
		return "not_found"
	case fiber.StatusInternalServerError:
		return "internal_error"
	case fiber.StatusServiceUnavailable:
		return "service_unavailable"
	default:
		return "error"
	}
}

// ValidationErrorResponse returns a 400 Bad Request with a validation error.
func ValidationErrorResponse(c fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
		Error:   "validation_error",
		Message: msg,
		Code:    fiber.StatusBadRequest,
	})
}

// AuthErrorResponse returns a 401 Unauthorized response.
func AuthErrorResponse(c fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
		Error:   "unauthorized",
		Message: msg,
		Code:    fiber.StatusUnauthorized,
	})
}

// NotFoundResponse returns a 404 Not Found response.
func NotFoundResponse(c fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
		Error:   "not_found",
		Message: msg,
		Code:    fiber.StatusNotFound,
	})
}

// InternalErrorResponse returns a 500 Internal Server Error response.
func InternalErrorResponse(c fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		Error:   "internal_error",
		Message: msg,
		Code:    fiber.StatusInternalServerError,
	})
}
