package signup

import (
	"context"
	"fmt"

	hashpassword "dvith.com/go-service-api/internal/security/hash_password"
	"dvith.com/go-service-api/internal/security/token"
)

// SignupRequest represents the user signup request
type SignupRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=255"`
	FullName string `json:"full_name" validate:"required,max=255"`
	Username string `json:"username" validate:"required,min=3,max=100"`
}

// SignupResponse represents the signup response with user and tokens
type SignupResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// SignupService handles user signup operations
type SignupService struct {
	repo         *SignupRepository
	tokenManager *token.TokenManager
}

// NewSignupService creates a new signup service with token manager
func NewSignupService(repo *SignupRepository, tokenManager *token.TokenManager) *SignupService {
	return &SignupService{
		repo:         repo,
		tokenManager: tokenManager,
	}
}

// RegisterUser registers a new user with password hashing and returns tokens
func (s *SignupService) RegisterUser(ctx context.Context, req *SignupRequest) (*SignupResponse, error) {
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

	// Generate JWT tokens
	tokenPair, err := s.tokenManager.GenerateTokenPair(savedUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &SignupResponse{
		User:         savedUser,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}
