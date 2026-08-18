package jwt

import (
	"fmt"

	"github.com/MiRRoRise/chat-service/internal/domain"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

type TokenManager interface {
	ValidateAccessToken(tokenString string) (int64, error)
}

type Manager struct {
	secretKey []byte
}

func NewManager(secret string) *Manager {
	return &Manager{secretKey: []byte(secret)}
}

func (m *Manager) ValidateAccessToken(tokenString string) (int64, error) {
	return m.validate(tokenString, "access")
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
