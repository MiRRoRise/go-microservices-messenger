package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/MiRRoRise/chat-service/internal/domain"
	"github.com/MiRRoRise/chat-service/internal/repository"
)

type ChatUseCase interface {
	CreateChat(ctx context.Context, name string) (*domain.Chat, error)
	GetChatByID(ctx context.Context, chatID int64) (*domain.Chat, error)
	ListChats(ctx context.Context) ([]domain.Chat, error)
}

type chatUseCase struct {
	chatRepo repository.ChatRepository
}

func NewChatUseCase(chatRepo repository.ChatRepository) *chatUseCase {
	return &chatUseCase{
		chatRepo: chatRepo,
	}
}

var _ ChatUseCase = (*chatUseCase)(nil)

func (u *chatUseCase) CreateChat(ctx context.Context, name string) (*domain.Chat, error) {
	if name == "" {
		return nil, domain.ErrChatNameRequired
	}

	chat := &domain.Chat{
		Name: name,
	}

	if err := u.chatRepo.CreateChat(ctx, chat); err != nil {
		if errors.Is(err, domain.ErrChatExists) {
			return nil, err
		}
		return nil, fmt.Errorf("sql create chat error: %w", err)
	}

	return chat, nil
}

func (u *chatUseCase) GetChatByID(ctx context.Context, chatID int64) (*domain.Chat, error) {
	if chatID <= 0 {
		return nil, domain.ErrChatIDRequired
	}

	chat, err := u.chatRepo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, domain.ErrChatNotFound
	}

	return chat, nil
}

func (u *chatUseCase) ListChats(ctx context.Context) ([]domain.Chat, error) {
	chats, err := u.chatRepo.ListChats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list chats: %w", err)
	}
	return chats, nil
}
