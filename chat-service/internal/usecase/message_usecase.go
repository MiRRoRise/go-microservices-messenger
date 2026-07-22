package usecase

import (
	"context"
	"fmt"

	"github.com/MiRRoRise/chat-service/internal/domain"
	"github.com/MiRRoRise/chat-service/internal/repository"
)

type MessageUseCase interface {
	CreateMessage(ctx context.Context, chatID, senderID int64, text string) (*domain.Message, error)
	ListMessages(ctx context.Context, chatID int64) ([]domain.Message, error)
}

type messageUseCase struct {
	messageRepo repository.MessageRepository
	chatRepo repository.ChatRepository
}

func NewMessageUseCase(messageRepo repository.MessageRepository, chatRepo repository.ChatRepository) *messageUseCase {
	return &messageUseCase{
		messageRepo: messageRepo,
		chatRepo: chatRepo,
	}
}

var _ MessageUseCase = (*messageUseCase)(nil)

func (u *messageUseCase) CreateMessage(ctx context.Context, chatID, senderID int64, text string) (*domain.Message, error) {
	if text == "" {
		return nil, domain.ErrMessageEmpty
	}

	_, err := u.chatRepo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, domain.ErrChatNotFound
	}

	message := &domain.Message{
		ChatID: chatID,
		SenderID: senderID,
		Text: text,
	}

	if err := u.messageRepo.CreateMessage(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return message, nil
}

func (u *messageUseCase) ListMessages(ctx context.Context, chatID int64) ([]domain.Message, error) {
	_, err := u.chatRepo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, domain.ErrChatNotFound
	}

	messages, err := u.messageRepo.ListMessages(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to list meassages: %w", err)
	}

	return messages, nil
}