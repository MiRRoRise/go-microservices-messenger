package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/MiRRoRise/chat-service/internal/domain"
)

type PostgresChatRepository struct {
	db *sql.DB
}

func NewChatRepo(db *sql.DB) *PostgresChatRepository {
	return &PostgresChatRepository{
		db: db,
	}
}

var _ ChatRepository = (*PostgresChatRepository)(nil)

func (r *PostgresChatRepository) CreateChat(ctx context.Context, chat *domain.Chat) error {
	err := r.db.QueryRowContext(
		ctx,
		"INSERT INTO chats (name) VALUES($1) RETURNING id, created_at",
		chat.Name,
	).Scan(&chat.ID, &chat.CreatedAt)

	if err != nil {
		return fmt.Errorf("sql error create chat: %w", err)
	}

	return nil
}

func (r *PostgresChatRepository) GetChatByID(ctx context.Context, chatID int64) (*domain.Chat, error) {
	var chat domain.Chat

	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, name, created_at FROM chats WHERE id = $1",
		chatID,
	).Scan(&chat.ID, &chat.Name, &chat.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrChatNotFound
		}
		return nil, fmt.Errorf("sql error get chat: %w", err)
	}

	return &chat, nil
}

func (r *PostgresChatRepository) ListChats(ctx context.Context) ([]domain.Chat, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, name, created_at FROM chats ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list chats: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var chats []domain.Chat
	for rows.Next() {
		var chat domain.Chat
		if err := rows.Scan(&chat.ID, &chat.Name, &chat.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan chat: %w", err)
		}
		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return chats, nil
}
