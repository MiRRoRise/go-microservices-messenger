package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MiRRoRise/chat-service/internal/delivery/dto"
	"github.com/MiRRoRise/chat-service/internal/domain"
	"github.com/MiRRoRise/chat-service/pkg/jwt"
	"github.com/go-chi/chi/v5"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockChatUseCase struct {
	mock.Mock
}

func (m *MockChatUseCase) CreateChat(ctx context.Context, name string) (*domain.Chat, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Chat), args.Error(1)
}

func (m *MockChatUseCase) GetChatByID(ctx context.Context, chatID int64) (*domain.Chat, error) {
	args := m.Called(ctx, chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Chat), args.Error(1)
}

func (m *MockChatUseCase) ListChats(ctx context.Context) ([]domain.Chat, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Chat), args.Error(1)
}

type MockMessageUseCase struct {
	mock.Mock
}

func (m *MockMessageUseCase) CreateMessage(ctx context.Context, chatID, senderID int64, text string) (*domain.Message, error) {
	args := m.Called(ctx, chatID, senderID, text)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Message), args.Error(1)
}

func (m *MockMessageUseCase) ListMessages(ctx context.Context, chatID int64) ([]domain.Message, error) {
	args := m.Called(ctx, chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Message), args.Error(1)
}

func TestHealthCheck(t *testing.T) {
	h := NewHandler(new(MockChatUseCase), new(MockMessageUseCase), jwt.NewManager("secret"))
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.HealthCheck(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
}

func TestCreateChat_Success(t *testing.T) {
	chatUC := new(MockChatUseCase)
	handler := NewHandler(chatUC, new(MockMessageUseCase), jwt.NewManager("secret"))

	chatUC.On("CreateChat", mock.Anything, "general").Return(&domain.Chat{
		ID:        1,
		Name:      "general",
		CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	}, nil)

	reqBody := `{"name":"general"}`
	req := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBufferString(reqBody))
	w := httptest.NewRecorder()

	handler.CreateChat(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp dto.CreateChatResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))

	assert.Equal(t, int64(1), resp.ID)
	assert.Equal(t, "general", resp.Name)
	chatUC.AssertExpectations(t)
}

func TestCreateChat_NameRequired(t *testing.T) {
	chatUC := new(MockChatUseCase)
	handler := NewHandler(chatUC, new(MockMessageUseCase), jwt.NewManager("secret"))

	chatUC.On("CreateChat", mock.Anything, "").Return(nil, domain.ErrChatNameRequired)

	req := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBufferString(`{"name":""}`))
	w := httptest.NewRecorder()

	handler.CreateChat(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateMessage_Success(t *testing.T) {
	msgUC := new(MockMessageUseCase)
	tokens := jwt.NewManager("secret")
	handler := NewHandler(new(MockChatUseCase), msgUC, tokens)

	msgUC.On("CreateMessage", mock.Anything, int64(5), int64(9), "hi").Return(&domain.Message{
		ID:        3,
		ChatID:    5,
		SenderID:  9,
		Text:      "hi",
		CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	}, nil)

	access, err := signAccessToken("secret", 9)
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Route("/chats/{id}/messages", func(r chi.Router) {
		r.Use(handler.AuthMiddleware)
		r.Post("/", handler.CreateMessage)
	})

	req := httptest.NewRequest(http.MethodPost, "/chats/5/messages", bytes.NewBufferString(`{"text":"hi"}`))
	req.Header.Set("Authorization", "Bearer "+access)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	msgUC.AssertExpectations(t)
}

func TestAuthMiddleware_RequiresAccessToken(t *testing.T) {
	tokens := jwt.NewManager("secret")
	handler := NewHandler(new(MockChatUseCase), new(MockMessageUseCase), tokens)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := GetIDFromContext(r.Context())
		require.NoError(t, err)
		assert.Equal(t, int64(123), id)
		w.WriteHeader(http.StatusNoContent)
	})

	accessToken, err := signAccessToken("secret", 123)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/chats", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w := httptest.NewRecorder()

	handler.AuthMiddleware(next).ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	refreshToken, err := signRefreshToken("secret", 123)
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodGet, "/chats", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	w = httptest.NewRecorder()

	handler.AuthMiddleware(next).ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func signAccessToken(secret string, userID int64) (string, error) {
	return signToken(secret, "access", userID, time.Hour)
}

func signRefreshToken(secret string, userID int64) (string, error) {
	return signToken(secret, "refresh", userID, time.Hour)
}

func signToken(secret, tokenType string, userID int64, ttl time.Duration) (string, error) {
	claims := jwtv5.MapClaims{
		"user_id": float64(userID),
		"exp":     time.Now().Add(ttl).Unix(),
		"type":    tokenType,
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
