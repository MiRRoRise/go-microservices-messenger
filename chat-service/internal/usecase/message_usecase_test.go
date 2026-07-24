package usecase

import (
	"context"
	"testing"

	"github.com/MiRRoRise/chat-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) CreateMessage(ctx context.Context, message *domain.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageRepository) ListMessages(ctx context.Context, chatID int64) ([]domain.Message, error) {
	args := m.Called(ctx, chatID)
	return args.Get(0).([]domain.Message), args.Error(1)
}

func TestMessageUseCase_CreateMessage_Success(t *testing.T) {
	mockChatRepo := new(MockChatRepository)
	mockMessageRepo := new(MockMessageRepository)

	u := NewMessageUseCase(mockMessageRepo, mockChatRepo, nil)
	ctx := context.Background()

	msg := &domain.Message{
		ChatID:   int64(123),
		SenderID: int64(2),
		Text:     "correct message",
	}

	mockChatRepo.On("GetChatByID", ctx, int64(123)).Return(&domain.Chat{ID: 123}, nil)
	mockMessageRepo.On("CreateMessage", ctx, msg).Return(nil)

	message, err := u.CreateMessage(ctx, int64(123), int64(2), "correct message")

	assert.NoError(t, err)
	assert.Equal(t, msg, message)
	mockChatRepo.AssertExpectations(t)
	mockMessageRepo.AssertExpectations(t)
}

func TestMessageUseCase_CreateMessage_EmptyText(t *testing.T) {
	mockChatRepo := new(MockChatRepository)
	mockMessageRepo := new(MockMessageRepository)

	u := NewMessageUseCase(mockMessageRepo, mockChatRepo, nil)
	ctx := context.Background()

	msg, err := u.CreateMessage(ctx, 1, 15, "")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrMessageEmpty, err)
	assert.Nil(t, msg)
}

func TestMessageUseCase_CreateMessage_ChatNotFound(t *testing.T) {
	mockChatRepo := new(MockChatRepository)
	mockMessageRepo := new(MockMessageRepository)

	u := NewMessageUseCase(mockMessageRepo, mockChatRepo, nil)
	ctx := context.Background()

	mockChatRepo.On("GetChatByID", ctx, int64(123)).Return(nil, domain.ErrChatNotFound)

	message, err := u.CreateMessage(ctx, int64(123), int64(2), "correct message")

	assert.Error(t, err)
	assert.EqualError(t, err, domain.ErrChatNotFound.Error())
	assert.Nil(t, message)
}

func TestMessageUseCase_ListMessages_Success(t *testing.T) {
	mockChatRepo := new(MockChatRepository)
	mockMessageRepo := new(MockMessageRepository)

	u := NewMessageUseCase(mockMessageRepo, mockChatRepo, nil)
	ctx := context.Background()

	expected := []domain.Message{
		{ID: 1, ChatID: 123, SenderID: 15, Text: "first message"},
		{ID: 2, ChatID: 123, SenderID: 15, Text: "second message"},
	}

	mockChatRepo.On("GetChatByID", ctx, int64(123)).Return(&domain.Chat{ID: 123}, nil)
	mockMessageRepo.On("ListMessages", ctx, int64(123)).Return(expected, nil)

	messages, err := u.ListMessages(ctx, int64(123))

	assert.NoError(t, err)
	assert.Equal(t, expected, messages)
	mockChatRepo.AssertExpectations(t)
	mockMessageRepo.AssertExpectations(t)
}

func TestMessageUseCase_ListMessages_ChatNotFound(t *testing.T) {
	mockChatRepo := new(MockChatRepository)
	mockMessageRepo := new(MockMessageRepository)

	u := NewMessageUseCase(mockMessageRepo, mockChatRepo, nil)
	ctx := context.Background()


	mockChatRepo.On("GetChatByID", ctx, int64(123)).Return(nil, domain.ErrChatNotFound)

	messages, err := u.ListMessages(ctx, int64(123))

	assert.Error(t, err)
	assert.Equal(t, domain.ErrChatNotFound, err)
	assert.Nil(t, messages)
}
