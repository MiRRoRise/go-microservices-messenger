package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/MiRRoRise/chat-service/internal/domain"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	connStr := "user=user password=password dbname=chat_db host=localhost port=5433 sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	if pingErr := db.Ping(); pingErr != nil {
		t.Skipf("postgres not available: %v", pingErr)
	}

	_, err = db.Exec("TRUNCATE TABLE chats RESTART IDENTITY CASCADE")
	require.NoError(t, err)
	_, err = db.Exec("TRUNCATE TABLE messages RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	return db
}

func TestChatRepository_CreateChat(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewChatRepo(db)
	ctx := context.Background()

	chat := &domain.Chat{Name: "Test chat"}

	err := repo.CreateChat(ctx, chat)

	assert.NoError(t, err)
	assert.NotZero(t, chat.ID)
	assert.NotZero(t, chat.CreatedAt)
}
