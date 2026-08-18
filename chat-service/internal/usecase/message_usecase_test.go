package usecase

import (
	"context"
	"testing"

	"github.com/MiRRoRise/chat-service/internal/domain"
	"github.com/MiRRoRise/chat-service/internal/kafka"
	pb "github.com/MiRRoRise/chat-service/proto/auth"
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

type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) GetUserByID(ctx context.Context, userID int64) (*pb.GetUserResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetUserResponse), args.Error(1)
}

func (m *MockAuthClient) ValidateUser(ctx context.Context, userID int64) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockKafkaProducer struct {
	mock.Mock
}

func (m *MockKafkaProducer) PublishMessageCreated(event kafka.MessageCreatedEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func TestMessageUseCase_CreateMessage_Success(t *testing.T) {
	mockChatRepo := new(MockChatRepository)
	mockMessageRepo := new(MockMessageRepository)
	mockAuth := new(MockAuthClient)
	mockKafka := new(MockKafkaProducer)

	u := NewMessageUseCase(mockMessageRepo, mockChatRepo, mockAuth, mockKafka, nil)
	ctx := context.Background()

	mockChatRepo.On("GetChatByID", ctx, int64(123)).Return(&domain.Chat{ID: 123}, nil)
	mockAuth.On("ValidateUser", ctx, int64(2)).Return(true, nil)
	mockMessageRepo.On("CreateMessage", ctx, mock.AnythingOfType("*domain.Message")).Return(nil).Run(func(args mock.Arguments) {
		msg := args.Get(1).(*domain.Message)
		msg.ID = 1
	})
	mockKafka.On("PublishMessageCreated", mock.AnythingOfType("kafka.MessageCreatedEvent")).Return(nil)

	message, err := u.CreateMessage(ctx, int64(123), int64(2), "correct message")

	assert.NoError(t, err)
	assert.Equal(t, "correct message", message.Text)
	assert.Equal(t, int64(123), message.ChatID)
	mockChatRepo.AssertExpectations(t)
	mockMessageRepo.AssertExpectations(t)
	mockAuth.AssertExpectations(t)
	mockKafka.AssertExpectations(t)
}

func TestMessageUseCase_CreateMessage_EmptyText(t *testing.T) {
	mockChatRepo := new(MockChatRepository)
	mockMessageRepo := new(MockMessageRepository)

	u := NewMessageUseCase(mockMessageRepo, mockChatRepo, nil, nil, nil)
	ctx := context.Background()

	msg, err := u.CreateMessage(ctx, 1, 15, "")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrMessageEmpty, err)
	assert.Nil(t, msg)
}

func TestMessageUseCase_CreateMessage_ChatNotFound(t *testing.T) {
	mockChatRepo := new(MockChatRepository)
	mockMessageRepo := new(MockMessageRepository)

	u := NewMessageUseCase(mockMessageRepo, mockChatRepo, nil, nil, nil)
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

	u := NewMessageUseCase(mockMessageRepo, mockChatRepo, nil, nil, nil)
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

	u := NewMessageUseCase(mockMessageRepo, mockChatRepo, nil, nil, nil)
	ctx := context.Background()

	mockChatRepo.On("GetChatByID", ctx, int64(123)).Return(nil, domain.ErrChatNotFound)

	messages, err := u.ListMessages(ctx, int64(123))

	assert.Error(t, err)
	assert.Equal(t, domain.ErrChatNotFound, err)
	assert.Nil(t, messages)
}
