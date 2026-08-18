package usecase

import (
	"context"
	"testing"

	"github.com/MiRRoRise/chat-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) CreateChat(ctx context.Context, chat *domain.Chat) error {
	args := m.Called(ctx, chat)
	return args.Error(0)
}

func (m *MockChatRepository) GetChatByID(ctx context.Context, chatID int64) (*domain.Chat, error) {
	args := m.Called(ctx, chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Chat), args.Error(1)
}

func (m *MockChatRepository) ListChats(ctx context.Context) ([]domain.Chat, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Chat), args.Error(1)
}

func TestChatUseCase_CreateChat_Success(t *testing.T) {
	mockRepo := new(MockChatRepository)
	u := NewChatUseCase(mockRepo, nil, nil)
	ctx := context.Background()

	mockRepo.On("CreateChat", ctx, mock.AnythingOfType("*domain.Chat")).Return(nil)

	chat, err := u.CreateChat(ctx, "test chat")

	assert.NoError(t, err)
	assert.Equal(t, "test chat", chat.Name)
	mockRepo.AssertExpectations(t)
}

func TestChatUseCase_CreateChat_Exists(t *testing.T) {
	mockRepo := new(MockChatRepository)
	u := NewChatUseCase(mockRepo, nil, nil)
	ctx := context.Background()

	mockRepo.On("CreateChat", ctx, mock.AnythingOfType("*domain.Chat")).Return(domain.ErrChatExists)

	chat, err := u.CreateChat(ctx, "chat exists")

	assert.Error(t, err)
	assert.EqualError(t, err, domain.ErrChatExists.Error())
	assert.Nil(t, chat)
	mockRepo.AssertExpectations(t)
}

func TestChatUseCase_GetChatByID_Success(t *testing.T) {
	mockRepo := new(MockChatRepository)
	u := NewChatUseCase(mockRepo, nil, nil)
	ctx := context.Background()

	existing := &domain.Chat{
		ID:   int64(123),
		Name: "exist",
	}

	mockRepo.On("GetChatByID", ctx, int64(123)).Return(existing, nil)

	chat, err := u.GetChatByID(ctx, int64(123))

	assert.NoError(t, err)
	assert.Equal(t, existing, chat)
	mockRepo.AssertExpectations(t)
}

func TestChatUseCase_GetChatByID_NotFound(t *testing.T) {
	mockRepo := new(MockChatRepository)
	u := NewChatUseCase(mockRepo, nil, nil)
	ctx := context.Background()

	mockRepo.On("GetChatByID", ctx, int64(1)).Return(nil, domain.ErrChatNotFound)

	chat, err := u.GetChatByID(ctx, 1)

	assert.Error(t, err)
	assert.EqualError(t, err, domain.ErrChatNotFound.Error())
	assert.Nil(t, chat)
	mockRepo.AssertExpectations(t)
}

func TestChatUseCase_ListChats_Success(t *testing.T) {
	mockRepo := new(MockChatRepository)
	u := NewChatUseCase(mockRepo, nil, nil)
	ctx := context.Background()

	expected := []domain.Chat{
		{ID: 1, Name: "chat1"},
		{ID: 2, Name: "chat2"},
	}

	mockRepo.On("ListChats", ctx).Return(expected, nil)

	chats, err := u.ListChats(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expected, chats)
	mockRepo.AssertExpectations(t)
}
