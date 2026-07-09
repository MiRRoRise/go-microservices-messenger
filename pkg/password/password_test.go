package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBcryptHasher_Hash(t *testing.T) {
	hasher := NewBcryptHasher(4)

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: "very_long_password_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := hasher.Hash(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.NotEqual(t, tt.password, hash)
		})
	}
}

func TestBcryptHasher_Compare(t *testing.T) {
	hasher := NewBcryptHasher(4)

	t.Run("correct password", func(t *testing.T) {
		password := "correct"
		hash, err := hasher.Hash(password)
		require.NoError(t, err)

		err = hasher.Compare(hash, password)
		assert.NoError(t, err)
	})

	t.Run("wrong password", func(t *testing.T) {
		password := "password123"
		hash, err := hasher.Hash(password)
		assert.NoError(t, err)

		err = hasher.Compare(hash, "wrong_password")
		assert.Error(t, err)
	})

	t.Run("incorrect hash", func(t *testing.T) {
		err := hasher.Compare("incorrect_hash", "password")
		assert.Error(t, err)
	})
}
