package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenConfig holds JWT configuration
type TokenConfig struct {
	SecretKey       string        // Secret key for signing tokens
	ExpirationTime  time.Duration // Token expiration duration
	RefreshDuration time.Duration // Refresh token expiration duration
	Issuer          string        // JWT issuer claim
}

// Claims represents custom JWT claims
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// RefreshTokenClaims represents refresh token claims
type RefreshTokenClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// TokenManager handles JWT token operations
type TokenManager struct {
	config TokenConfig
}

// NewTokenManager creates a new token manager
func NewTokenManager(config TokenConfig) *TokenManager {
	return &TokenManager{
		config: config,
	}
}

// GenerateTokenPair generates both access and refresh tokens
func (tm *TokenManager) GenerateTokenPair(userID uuid.UUID) (*TokenPair, error) {
	// Generate access token
	accessToken, err := tm.GenerateAccessToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := tm.GenerateRefreshToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(tm.config.ExpirationTime.Seconds()),
	}, nil
}

// GenerateAccessToken generates a JWT access token
func (tm *TokenManager) GenerateAccessToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	expirationTime := now.Add(tm.config.ExpirationTime)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    tm.config.Issuer,
			Audience:  jwt.ClaimStrings{tm.config.Issuer + "-users"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tm.config.SecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a JWT refresh token
func (tm *TokenManager) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	expirationTime := now.Add(tm.config.RefreshDuration)

	claims := &RefreshTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    tm.config.Issuer,
			Audience:  jwt.ClaimStrings{tm.config.Issuer + "-refresh"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tm.config.SecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

// ValidateAccessToken validates and parses an access token
func (tm *TokenManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tm.config.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse access token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid access token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Verify issuer and audience
	if claims.Issuer != tm.config.Issuer {
		return nil, fmt.Errorf("invalid token issuer")
	}

	// Check if audience contains our expected value
	found := false
	for _, aud := range claims.Audience {
		if aud == "go-service-api-users" {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("invalid token audience")
	}

	return claims, nil
}

// ValidateRefreshToken validates and parses a refresh token
func (tm *TokenManager) ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tm.config.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Verify issuer and audience
	if claims.Issuer != tm.config.Issuer {
		return nil, fmt.Errorf("invalid token issuer")
	}

	// Check if audience contains our expected value
	found := false
	for _, aud := range claims.Audience {
		if aud == "go-service-api-refresh" {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("invalid token audience")
	}

	return claims, nil
}
