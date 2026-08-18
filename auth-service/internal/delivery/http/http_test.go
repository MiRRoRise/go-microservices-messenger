package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MiRRoRise/auth-service/internal/delivery/dto"
	"github.com/MiRRoRise/auth-service/internal/domain"
	"github.com/MiRRoRise/auth-service/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockUserUseCase struct {
	mock.Mock
}

func (m *MockUserUseCase) Register(ctx context.Context, email, password string) (*domain.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserUseCase) Login(ctx context.Context, email, password string) (string, string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockUserUseCase) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	args := m.Called(ctx, refreshToken)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockUserUseCase) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func TestHandler_Register_Success(t *testing.T) {
	mockUC := new(MockUserUseCase)
	handler := NewHandler(mockUC, jwt.NewManager("secret"))

	reqBody := `{"email":"user@gmail.com", "password":"password123"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(reqBody))
	w := httptest.NewRecorder()

	mockUC.On("Register", mock.Anything, "user@gmail.com", "password123").Return(&domain.User{ID: int64(123), Email: "user@gmail.com", PasswordHash: "hashed_password"}, nil)

	handler.Register(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp dto.RegisterResponse
	err := json.NewDecoder(w.Body).Decode(&resp)

	assert.NoError(t, err)
	assert.Equal(t, int64(123), resp.ID)
	assert.Equal(t, "user@gmail.com", resp.Email)
	mockUC.AssertExpectations(t)
}

func TestHandler_Register_InvalidJSON(t *testing.T) {
	mockUC := new(MockUserUseCase)
	handler := NewHandler(mockUC, jwt.NewManager("secret"))

	reqBody := `{email: "wrong@gmail.com", "password":"password"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(reqBody))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthMiddleware_AccessToken(t *testing.T) {
	mockUC := new(MockUserUseCase)
	tokens := jwt.NewManager("secret")
	handler := NewHandler(mockUC, tokens)

	access, err := tokens.GenerateAccessToken(123)
	require.NoError(t, err)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		id, err := GetIDFromContext(r.Context())
		require.NoError(t, err)
		assert.Equal(t, int64(123), id)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	w := httptest.NewRecorder()

	handler.AuthMiddleware(next).ServeHTTP(w, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_RefreshTokenRejected(t *testing.T) {
	mockUC := new(MockUserUseCase)
	tokens := jwt.NewManager("secret")
	handler := NewHandler(mockUC, tokens)

	refresh, err := tokens.GenerateRefreshToken(7)
	require.NoError(t, err)

	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("next should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+refresh)
	w := httptest.NewRecorder()

	handler.AuthMiddleware(next).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	handler := NewHandler(new(MockUserUseCase), jwt.NewManager("secret"))
	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("next should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()
	handler.AuthMiddleware(next).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
