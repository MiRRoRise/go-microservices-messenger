package repository

import (
	"context"

	"github.com/MiRRoRise/chat-service/internal/domain"
)

type ChatRepository interface {
	CreateChat(ctx context.Context, chat *domain.Chat) error
	GetChatByID(ctx context.Context, chatID int64) (*domain.Chat, error)
	ListChats(ctx context.Context) ([]domain.Chat, error)
}

type MessageRepository interface {
	CreateMessage(ctx context.Context, message *domain.Message) error
	ListMessages(ctx context.Context, chatID int64) ([]domain.Message, error)
}
