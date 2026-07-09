package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_GenerateAccessToken(t *testing.T) {
	manager := NewManager("secret")

	token, err := manager.GenerateAccessToken(123)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestManager_GenerateRefreshToken(t *testing.T) {
	manager := NewManager("secret")

	token, err := manager.GenerateRefreshToken(123)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	userID, err := manager.ValidateRefreshToken(token)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), userID)
}

func TestManager_ValidateRefreshToken_WrongType(t *testing.T) {
	manager := NewManager("secret")

	token, err := manager.GenerateAccessToken(123)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	_, err = manager.ValidateRefreshToken(token)
	assert.Error(t, err)
}

func TestManager_ValidateRefreshToken_Expired(t *testing.T) {
	manager := NewManager("secret")

	claims := jwt.MapClaims{
		"user_id": int64(123),
		"exp":     time.Now().Add(-1 * time.Minute).Unix(),
		"type":    "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expired, err := token.SignedString([]byte("secret"))
	require.NoError(t, err)

	_, err = manager.ValidateRefreshToken(expired)
	assert.Error(t, err)
}

func TestManager_ValidateRefreshToken_WrongSecret(t *testing.T) {
	manager1 := NewManager("secret1")
	manager2 := NewManager("secret2")

	token, err := manager1.GenerateAccessToken(123)
	require.NoError(t, err)

	_, err = manager2.ValidateRefreshToken(token)
	assert.Error(t, err)
}

func TestManager_ValidateRefreshToken_InvalidFormat(t *testing.T) {
	manager := NewManager("secret")

	_, err := manager.ValidateRefreshToken("invalid.token")
	assert.Error(t, err)
}