package jwt_test

import (
	"testing"
	"time"

	"github.com/MiRRoRise/auth-service/internal/domain"
	"github.com/MiRRoRise/auth-service/pkg/jwt"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_AccessAndRefresh(t *testing.T) {
	m := jwt.NewManager("secret")

	access, err := m.GenerateAccessToken(42)
	require.NoError(t, err)

	id, err := m.ValidateAccessToken(access)
	require.NoError(t, err)
	assert.Equal(t, int64(42), id)

	_, err = m.ValidateRefreshToken(access)
	assert.ErrorIs(t, err, domain.ErrInvalidToken)

	refresh, err := m.GenerateRefreshToken(42)
	require.NoError(t, err)

	id, err = m.ValidateRefreshToken(refresh)
	require.NoError(t, err)
	assert.Equal(t, int64(42), id)

	_, err = m.ValidateAccessToken(refresh)
	assert.ErrorIs(t, err, domain.ErrInvalidToken)
}

func TestManager_Expired(t *testing.T) {
	m := jwt.NewManager("secret")
	claims := jwtv5.MapClaims{
		"user_id": float64(1),
		"exp":     time.Now().Add(-time.Minute).Unix(),
		"type":    "access",
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	raw, err := token.SignedString([]byte("secret"))
	require.NoError(t, err)

	_, err = m.ValidateAccessToken(raw)
	assert.ErrorIs(t, err, domain.ErrInvalidToken)
}
