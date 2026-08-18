package jwt

import (
	"fmt"
	"time"

	"github.com/MiRRoRise/auth-service/internal/domain"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

type TokenManager interface {
	GenerateAccessToken(userID int64) (string, error)
	GenerateRefreshToken(userID int64) (string, error)
	ValidateAccessToken(tokenString string) (int64, error)
	ValidateRefreshToken(tokenString string) (int64, error)
}

type Manager struct {
	secretKey []byte
}

func NewManager(secret string) *Manager {
	return &Manager{secretKey: []byte(secret)}
}

func (m *Manager) GenerateAccessToken(userID int64) (string, error) {
	return m.generate(userID, "access", 15*time.Minute)
}

func (m *Manager) GenerateRefreshToken(userID int64) (string, error) {
	return m.generate(userID, "refresh", 7*24*time.Hour)
}

func (m *Manager) generate(userID int64, tokenType string, ttl time.Duration) (string, error) {
	claims := jwtv5.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(ttl).Unix(),
		"iat":     time.Now().Unix(),
		"type":    tokenType,
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *Manager) ValidateAccessToken(tokenString string) (int64, error) {
	return m.validate(tokenString, "access")
}

func (m *Manager) ValidateRefreshToken(tokenString string) (int64, error) {
	return m.validate(tokenString, "refresh")
}

func (m *Manager) validate(tokenString, expectedType string) (int64, error) {
	token, err := jwtv5.Parse(tokenString, func(t *jwtv5.Token) (any, error) {
		if _, ok := t.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secretKey, nil
	})
	if err != nil || !token.Valid {
		return 0, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwtv5.MapClaims)
	if !ok {
		return 0, domain.ErrInvalidToken
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != expectedType {
		return 0, domain.ErrInvalidToken
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, domain.ErrInvalidToken
	}

	return int64(userIDFloat), nil
}
