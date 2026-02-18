package token

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGenerateAccessToken(t *testing.T) {
	config := TokenConfig{
		SecretKey:      "test-secret-key",
		ExpirationTime: 1 * time.Hour,
		Issuer:         "go-service-api",
	}
	tm := NewTokenManager(config)

	userID := uuid.New()
	token, err := tm.GenerateAccessToken(userID)

	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	if token == "" {
		t.Errorf("GenerateAccessToken() returned empty token")
	}

	// Validate the token
	claims, err := tm.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID mismatch")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	config := TokenConfig{
		SecretKey:       "test-secret-key",
		RefreshDuration: 7 * 24 * time.Hour,
		Issuer:          "go-service-api",
	}
	tm := NewTokenManager(config)

	userID := uuid.New()
	token, err := tm.GenerateRefreshToken(userID)

	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	claims, err := tm.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("ValidateRefreshToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID mismatch")
	}
}

func TestGenerateTokenPair(t *testing.T) {
	config := TokenConfig{
		SecretKey:       "test-secret-key",
		ExpirationTime:  1 * time.Hour,
		RefreshDuration: 7 * 24 * time.Hour,
		Issuer:          "go-service-api",
	}
	tm := NewTokenManager(config)

	userID := uuid.New()
	pair, err := tm.GenerateTokenPair(userID)

	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}

	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Errorf("Token pair contains empty tokens")
	}

	if pair.TokenType != "Bearer" {
		t.Errorf("TokenType = %v, want Bearer", pair.TokenType)
	}
}

func TestValidateAccessToken_InvalidToken(t *testing.T) {
	config := TokenConfig{
		SecretKey:      "test-secret-key",
		ExpirationTime: 1 * time.Hour,
		Issuer:         "go-service-api",
	}
	tm := NewTokenManager(config)

	_, err := tm.ValidateAccessToken("invalid-token")
	if err == nil {
		t.Errorf("ValidateAccessToken() should error on invalid token")
	}
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	config1 := TokenConfig{
		SecretKey:      "secret-1",
		ExpirationTime: 1 * time.Hour,
		Issuer:         "go-service-api",
	}
	tm1 := NewTokenManager(config1)

	userID := uuid.New()
	token, _ := tm1.GenerateAccessToken(userID)

	config2 := TokenConfig{
		SecretKey:      "secret-2",
		ExpirationTime: 1 * time.Hour,
		Issuer:         "go-service-api",
	}
	tm2 := NewTokenManager(config2)

	_, err := tm2.ValidateAccessToken(token)
	if err == nil {
		t.Errorf("ValidateAccessToken() should error with wrong secret")
	}
}

func BenchmarkGenerateAccessToken(b *testing.B) {
	config := TokenConfig{
		SecretKey:      "test-secret-key",
		ExpirationTime: 1 * time.Hour,
		Issuer:         "go-service-api",
	}
	tm := NewTokenManager(config)
	userID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.GenerateAccessToken(userID)
	}
}

func BenchmarkValidateAccessToken(b *testing.B) {
	config := TokenConfig{
		SecretKey:      "test-secret-key",
		ExpirationTime: 1 * time.Hour,
		Issuer:         "go-service-api",
	}
	tm := NewTokenManager(config)
	token, _ := tm.GenerateAccessToken(uuid.New())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.ValidateAccessToken(token)
	}
}
