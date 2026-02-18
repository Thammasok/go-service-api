package signup

import (
	"context"
	"fmt"

	hashpassword "dvith.com/go-service-api/internal/security/hash_password"
)

// SignupRequest represents the user signup request
type SignupRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=255"`
	FullName string `json:"full_name" validate:"required,max=255"`
	Username string `json:"username" validate:"required,min=3,max=100"`
}

// SignupService handles user signup operations
type SignupService struct {
	repo *SignupRepository
}

// NewSignupService creates a new signup service
func NewSignupService(repo *SignupRepository) *SignupService {
	return &SignupService{
		repo: repo,
	}
}

// RegisterUser registers a new user with password hashing
func (s *SignupService) RegisterUser(ctx context.Context, req *SignupRequest) (*User, error) {
	if req == nil {
		return nil, fmt.Errorf("signup request cannot be nil")
	}

	// Validate password strength
	strength := ValidatePasswordStrength(req.Password)
	if !strength.IsValid {
		return nil, fmt.Errorf("password must contain uppercase letters, lowercase letters, numbers, and special characters")
	}

	// Hash the password
	hashedPassword, err := hashpassword.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user object
	user := &User{
		Email:    req.Email,
		Password: hashedPassword,
		FullName: req.FullName,
		Username: req.Username,
		IsActive: true,
	}

	// Save user to database
	savedUser, err := s.repo.SaveUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	return savedUser, nil
}
