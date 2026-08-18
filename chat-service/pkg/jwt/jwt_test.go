package jwt_test

import (
	"testing"
	"time"

	"github.com/MiRRoRise/chat-service/internal/domain"
	"github.com/MiRRoRise/chat-service/pkg/jwt"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_ValidateAccessToken(t *testing.T) {
	m := jwt.NewManager("secret")

	token, err := signToken("secret", "access", 42, time.Hour)
	require.NoError(t, err)

	id, err := m.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, int64(42), id)
}

func TestManager_RejectsRefreshToken(t *testing.T) {
	m := jwt.NewManager("secret")

	token, err := signToken("secret", "refresh", 42, time.Hour)
	require.NoError(t, err)

	_, err = m.ValidateAccessToken(token)
	assert.ErrorIs(t, err, domain.ErrInvalidToken)
}

func signToken(secret, tokenType string, userID int64, ttl time.Duration) (string, error) {
	claims := jwtv5.MapClaims{
		"user_id": float64(userID),
		"exp":     time.Now().Add(ttl).Unix(),
		"type":    tokenType,
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
