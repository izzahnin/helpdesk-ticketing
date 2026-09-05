package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"helpdesk-backend/internal/models"
)

func TestHashPasswordAndCheckPassword(t *testing.T) {
	svc := &AuthService{}

	hash, err := svc.HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword() returned empty hash")
	}
	if hash == "password123" {
		t.Fatal("HashPassword() returned the plain password")
	}
	if !svc.CheckPassword(hash, "password123") {
		t.Fatal("CheckPassword() returned false for the correct password")
	}
	if svc.CheckPassword(hash, "wrong-password") {
		t.Fatal("CheckPassword() returned true for the wrong password")
	}
}

func TestAccessToken(t *testing.T) {
	svc := &AuthService{AccessSecret: "test-secret"}
	user := &models.User{ID: 42, Role: "super_admin"}

	tokenString, err := svc.AccessToken(user)
	if err != nil {
		t.Fatalf("AccessToken() error = %v", err)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(svc.AccessSecret), nil
	})
	if err != nil {
		t.Fatalf("jwt.Parse() error = %v", err)
	}
	if !token.Valid {
		t.Fatal("parsed access token is invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("token claims are not jwt.MapClaims")
	}
	if got := claims["sub"]; got != float64(user.ID) {
		t.Fatalf("sub claim = %v, want %d", got, user.ID)
	}
	if got := claims["role"]; got != user.Role {
		t.Fatalf("role claim = %v, want %q", got, user.Role)
	}

	expFloat, ok := claims["exp"].(float64)
	if !ok {
		t.Fatalf("exp claim has type %T, want float64", claims["exp"])
	}
	expiresAt := time.Unix(int64(expFloat), 0)
	if time.Until(expiresAt) <= 14*time.Minute || time.Until(expiresAt) > 15*time.Minute {
		t.Fatalf("token expiry = %v, want around 15 minutes from now", expiresAt)
	}
}

func TestHashRefreshToken(t *testing.T) {
	svc := &AuthService{Pepper: "pepper-a"}

	hash := svc.HashRefreshToken("refresh-token")
	if hash == "" {
		t.Fatal("HashRefreshToken() returned empty hash")
	}
	if hash == "refresh-token" {
		t.Fatal("HashRefreshToken() returned plain token")
	}
	if got := svc.HashRefreshToken("refresh-token"); got != hash {
		t.Fatal("HashRefreshToken() should be deterministic for same input and pepper")
	}

	otherPepper := (&AuthService{Pepper: "pepper-b"}).HashRefreshToken("refresh-token")
	if otherPepper == hash {
		t.Fatal("HashRefreshToken() should change when pepper changes")
	}
}

func TestNewRefreshToken(t *testing.T) {
	svc := &AuthService{Pepper: "test-pepper"}

	plain, hash, expiresAt, err := svc.NewRefreshToken()
	if err != nil {
		t.Fatalf("NewRefreshToken() error = %v", err)
	}
	if plain == "" {
		t.Fatal("NewRefreshToken() returned empty plain token")
	}
	if hash == "" {
		t.Fatal("NewRefreshToken() returned empty hash")
	}
	if plain == hash {
		t.Fatal("NewRefreshToken() hash must not equal plain token")
	}
	if hash != svc.HashRefreshToken(plain) {
		t.Fatal("NewRefreshToken() hash does not match HashRefreshToken(plain)")
	}
	if time.Until(expiresAt) <= 6*24*time.Hour || time.Until(expiresAt) > 7*24*time.Hour {
		t.Fatalf("refresh token expiry = %v, want around 7 days from now", expiresAt)
	}
}
