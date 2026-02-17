package examples

import (
	"dvith.com/go-service-api/internal/middleware"
	"github.com/gofiber/fiber/v3"
)

// ExampleUserRequest represents a user creation request.
type ExampleUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ExampleCreateUserHandler shows how to use ValidationErrorResponse.
// POST /api/v1/examples/users
func ExampleCreateUserHandler(c fiber.Ctx) error {
	var req ExampleUserRequest

	// Parse request body using Fiber v3 API
	if err := c.Bind().JSON(&req); err != nil {
		// Return a validation error response (400)
		return middleware.ValidationErrorResponse(c, "invalid request body: "+err.Error())
	}

	// Validate required fields
	if req.Name == "" {
		return middleware.ValidationErrorResponse(c, "name is required")
	}
	if req.Email == "" {
		return middleware.ValidationErrorResponse(c, "email is required")
	}

	// Success response
	return c.JSON(fiber.Map{
		"id":    123,
		"name":  req.Name,
		"email": req.Email,
	})
}

// ExampleGetUserHandler shows how to use NotFoundResponse.
// GET /api/v1/examples/users/:id
func ExampleGetUserHandler(c fiber.Ctx) error {
	userID := c.Params("id")

	// Simulate user lookup
	if userID != "123" {
		// Return a not found error (404)
		return middleware.NotFoundResponse(c, "user with id "+userID+" not found")
	}

	return c.JSON(fiber.Map{
		"id":    123,
		"name":  "John Doe",
		"email": "john@example.com",
	})
}

// ExampleAuthenticatedHandler shows how to use AuthErrorResponse.
// GET /api/v1/examples/protected
func ExampleAuthenticatedHandler(c fiber.Ctx) error {
	// Get token from Authorization header
	token := c.Get("Authorization")

	// Check if token is valid
	if token == "" {
		return middleware.AuthErrorResponse(c, "missing authorization header")
	}

	if !isValidToken(token) {
		// Return an authorization error (401)
		return middleware.AuthErrorResponse(c, "invalid or expired token")
	}

	return c.JSON(fiber.Map{
		"message": "authenticated access granted",
		"user":    "john_doe",
		"token":   token,
	})
}

// ExampleDatabaseErrorHandler shows how to use InternalErrorResponse.
// GET /api/v1/examples/data
func ExampleDatabaseErrorHandler(c fiber.Ctx) error {
	// Simulate a database operation
	data, err := simulateDatabaseOperation()

	if err != nil {
		// Return an internal error response (500)
		// Don't expose internal details to client
		return middleware.InternalErrorResponse(c, "failed to retrieve data from database")
	}

	return c.JSON(fiber.Map{
		"data": data,
	})
}

// ExamplePanicHandler shows automatic panic recovery by ErrorHandler middleware.
// GET /api/v1/examples/panic
func ExamplePanicHandler(c fiber.Ctx) error {
	// This panic will be caught by the ErrorHandler middleware
	// and converted to a 500 response automatically with error logging
	panic("something went terribly wrong!")
}

// Helper functions

func isValidToken(token string) bool {
	// Simple token validation example
	// In real app, validate against JWT or session store
	validTokens := map[string]bool{
		"Bearer valid-token-123": true,
		"Bearer test-token":      true,
	}
	return validTokens[token]
}

func simulateDatabaseOperation() ([]string, error) {
	// Simulate a successful database operation
	return []string{"item1", "item2", "item3"}, nil

	// Uncomment to test error handling:
	// return nil, errors.New("database connection timeout")
}
