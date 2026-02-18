package signin

import (
	"context"
	"fmt"

	hashpassword "dvith.com/go-service-api/internal/security/hash_password"
	"dvith.com/go-service-api/internal/security/token"
)

type SigninRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SigninResponse represents the signin response with user and tokens
type SigninResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// SigninService handles user signin operations
type SigninService struct {
	repo         *SigninRepository
	tokenManager *token.TokenManager
}

// NewSigninService creates a new signin service with token manager
func NewSigninService(repo *SigninRepository, tokenManager *token.TokenManager) *SigninService {
	return &SigninService{
		repo:         repo,
		tokenManager: tokenManager,
	}
}

// LoginUser logs in a user with password hashing and returns tokens
func (s *SigninService) LoginUser(ctx context.Context, req *SigninRequest) (*SigninResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("signin request cannot be nil")
	}

	// Find user with email
	user, err := s.repo.FindUser(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to login user: %w", err)
	}

	// Check the password matches
	isPasswordMatch := hashpassword.CheckPassword(req.Password, user.Password)
	if isPasswordMatch == false {
		return nil, fmt.Errorf("login failed please recheck the username and password and try again")
	}

	// Generate JWT tokens
	tokenPair, err := s.tokenManager.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &SigninResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}
