package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MiRRoRise/chat-service/internal/domain"
)

type PostgresMessageRepository struct {
	db *sql.DB
}

func NewMessageRepo(db *sql.DB) *PostgresMessageRepository {
	return &PostgresMessageRepository{
		db: db,
	}
}

var _ MessageRepository = (*PostgresMessageRepository)(nil)

func (r *PostgresMessageRepository) CreateMessage(ctx context.Context, message *domain.Message) error {
	err := r.db.QueryRowContext(
		ctx,
		"INSERT INTO messages (chat_id, sender_id, text) VALUES($1, $2, $3) RETURNING id, created_at",
		message.ChatID, 
		message.SenderID, 
		message.Text,
	).Scan(&message.ID, &message.CreatedAt)

	if err != nil {
		return fmt.Errorf("sql error create message: %w", err)
	}

	return nil
}

func (r *PostgresMessageRepository) ListMessages(ctx context.Context, chatID int64) ([]domain.Message, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, chat_id, sender_id, text, created_at FROM messages WHERE chat_id = $1 ORDER BY created_at ASC",
		chatID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var message domain.Message
		if err := rows.Scan(&message.ID, &message.ChatID, &message.SenderID, &message.Text, &message.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return messages, nil
}