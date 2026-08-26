package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"helpdesk-backend/internal/models"
	"helpdesk-backend/internal/repository"
)

type AuthService struct {
	Users *repository.UserRepository
	RefreshRepo *repository.RefreshTokenRepository
	AccessSecret string
	Pepper string
}

func (s *AuthService) HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func (s *AuthService) CheckPassword(hash, password string) bool  {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (s *AuthService) AccessToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.ID,
		"role": user.Role,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.AccessSecret))
}

func (s *AuthService) NewRefreshToken() (plain string, hash string, expiresAt time.Time, err error) {
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return "", "", time.Time{}, err
	}
	plain = base64.RawURLEncoding.EncodeToString(raw)
	hash = s.HashRefreshToken(plain)
	expiresAt = time.Now().Add(7 * 24 * time.Hour)
	return plain, hash, expiresAt, nil
}

func (s *AuthService) HashRefreshToken(plain string) string {
	sum := sha256.Sum256([]byte(plain + s.Pepper))
	return hex.EncodeToString(sum[:])
}