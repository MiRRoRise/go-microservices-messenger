package jwt

import (
	"fmt"
	"time"

	"github.com/MiRRoRise/chat-service/internal/domain"
	"github.com/golang-jwt/jwt/v5"
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
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *Manager) GenerateRefreshToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *Manager) validateToken(tokenString string, expectedType string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("error validate token: %v", t.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		return 0, domain.ErrInvalidToken
	}

	if !token.Valid {
		return 0, fmt.Errorf("token invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid claims")
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != expectedType {
		return 0, fmt.Errorf("invalid token type")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("user id not found")
	}

	return int64(userIDFloat), nil
}

func (m *Manager) ValidateAccessToken(tokenString string) (int64, error) {
	return m.validateToken(tokenString, "access")
}

func (m *Manager) ValidateRefreshToken(tokenString string) (int64, error) {
	return m.validateToken(tokenString, "refresh")
}

