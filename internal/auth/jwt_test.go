package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndParseAccessToken(t *testing.T) {
	secret := []byte("my-super-secret-key")
	userID := "550e8400-e29b-41d4-a716-446655440000"

	// Generate token
	tokenStr, err := GenerateAccessToken(userID, secret)
	if err != nil {
		t.Fatalf("expected no error during token generation, got: %v", err)
	}

	if tokenStr == "" {
		t.Fatal("expected non-empty token string")
	}

	// Parse token
	claims, err := ParseAccessToken(tokenStr, secret)
	if err != nil {
		t.Fatalf("expected no error during token parsing, got: %v", err)
	}

	if claims.Subject != userID {
		t.Errorf("expected Subject %q, got %q", userID, claims.Subject)
	}
}

func TestParseAccessToken_InvalidSignature(t *testing.T) {
	secret1 := []byte("secret-one")
	secret2 := []byte("secret-two")
	userID := "user-123"

	tokenStr, _ := GenerateAccessToken(userID, secret1)

	// Attempt parsing with wrong secret
	_, err := ParseAccessToken(tokenStr, secret2)
	if err == nil {
		t.Fatal("expected parsing to fail with wrong secret, but it succeeded")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestParseAccessToken_ExpiredToken(t *testing.T) {
	secret := []byte("secret")
	userID := "user-123"

	// Create a token that has already expired
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Minute)), // Expired 1 min ago
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = ParseAccessToken(tokenStr, secret)
	if err == nil {
		t.Fatal("expected parsing of expired token to fail, but it succeeded")
	}
	if !errors.Is(err, ErrExpiredToken) {
		t.Errorf("expected ErrExpiredToken, got %v", err)
	}
}

func TestParseAccessToken_Malformed(t *testing.T) {
	secret := []byte("secret")
	_, err := ParseAccessToken("invalid.token.here", secret)
	if err == nil {
		t.Fatal("expected parsing of malformed token to fail")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}
