package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/MiRRoRise/auth-service/internal/domain"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	connStr := "user=user password=password dbname=db host=localhost port=5432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	if pingErr := db.Ping(); pingErr != nil {
		t.Skipf("postgres not available: %v", pingErr)
	}

	_, err = db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	return db
}

func TestUserRepository_CreateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()

	user := &domain.User{
		Email:        "qwerty@gmail.com",
		PasswordHash: "hashed_password",
	}

	err := repo.CreateUser(ctx, user)
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
}

func TestUserRepository_GetByEmail_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()

	user := &domain.User{
		Email:        "correct@email.com",
		PasswordHash: "hashed_password",
	}

	err := repo.CreateUser(ctx, user)
	require.NoError(t, err)

	userByEmail, err := repo.GetByEmail(ctx, user.Email)

	assert.NoError(t, err)
	assert.NotNil(t, userByEmail)
	assert.Equal(t, user.PasswordHash, userByEmail.PasswordHash)
	assert.Equal(t, user.ID, userByEmail.ID)
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()

	user, err := repo.GetByEmail(ctx, "nouser@gmail.com")

	assert.NoError(t, err)
	assert.Nil(t, user)
}

func TestUserRepository_GetByID_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()

	user := &domain.User{
		Email:        "correct@email.com",
		PasswordHash: "hashed_password",
	}

	err := repo.CreateUser(ctx, user)
	require.NoError(t, err)

	userByID, err := repo.GetByID(ctx, user.ID)

	assert.NoError(t, err)
	assert.NotNil(t, userByID)
	assert.Equal(t, user.Email, userByID.Email)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()

	user, err := repo.GetByID(ctx, 123)

	assert.NoError(t, err)
	assert.Nil(t, user)
}

func TestUserRepository_CreateUser_DuplicateEmaiL(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepo(db)
	ctx := context.Background()

	user1 := &domain.User{
		Email:        "email@email.com",
		PasswordHash: "hashed_password",
	}
	err := repo.CreateUser(ctx, user1)
	require.NoError(t, err)

	user2 := &domain.User{
		Email:        "email@email.com",
		PasswordHash: "another_hashed_password",
	}
	err = repo.CreateUser(ctx, user2)

	assert.Error(t, err)
}
