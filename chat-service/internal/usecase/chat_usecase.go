package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/MiRRoRise/chat-service/internal/domain"
	"github.com/MiRRoRise/chat-service/internal/repository"
	"github.com/MiRRoRise/chat-service/pkg/redis"
)

type ChatUseCase interface {
	CreateChat(ctx context.Context, name string) (*domain.Chat, error)
	GetChatByID(ctx context.Context, chatID int64) (*domain.Chat, error)
	ListChats(ctx context.Context) ([]domain.Chat, error)
}

type chatUseCase struct {
	chatRepo repository.ChatRepository
	redis    *redis.Client
}

func NewChatUseCase(chatRepo repository.ChatRepository, redis *redis.Client) *chatUseCase {
	return &chatUseCase{
		chatRepo: chatRepo,
		redis: redis,
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

	_ = u.redis.Del(ctx, "chats:list")

	return chat, nil
}

func (u *chatUseCase) GetChatByID(ctx context.Context, chatID int64) (*domain.Chat, error) {
	if chatID <= 0 {
		return nil, domain.ErrChatIDRequired
	}

	cacheKey := fmt.Sprintf("chat:%d", chatID)
	cached, err := u.redis.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var chat domain.Chat
		if err := json.Unmarshal([]byte(cached), &chat); err == nil {
			return &chat, nil
		}
	}

	chat, err := u.chatRepo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, domain.ErrChatNotFound
	}

	data, _ := json.Marshal(chat)
	_ = u.redis.Set(ctx, cacheKey, string(data), 5*time.Minute)

	return chat, nil
}

func (u *chatUseCase) ListChats(ctx context.Context) ([]domain.Chat, error) {
	cached, err := u.redis.Get(ctx, "chats:list")
	if err == nil && cached != "" {
		var chats []domain.Chat
		if err := json.Unmarshal([]byte(cached), &chats); err == nil {
			return chats, nil
		}
	}

	chats, err := u.chatRepo.ListChats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list chats: %w", err)
	}

	data, _ := json.Marshal(chats)
	_ = u.redis.Set(ctx, "chats:list", string(data), 2*time.Minute)

	return chats, nil
}
