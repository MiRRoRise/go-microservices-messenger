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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func (m *MockUserUseCase) ValidateRefreshToken(ctx context.Context, token string) (int64, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(int64), args.Error(1)
}

func TestHandler_Register_Success(t *testing.T) {
	mockUC := new(MockUserUseCase)
	handler := NewHandler(mockUC)

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
	handler := NewHandler(mockUC)

	reqBody := `{email: "wrong@gmail.com", "password":"password"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(reqBody))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Login_Success(t *testing.T) {
	mockUC := new(MockUserUseCase)
	handler := NewHandler(mockUC)

	reqBody := `{"email":"user@gmail.com", "password":"password123"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(reqBody))
	w := httptest.NewRecorder()

	mockUC.On("Login", mock.Anything, "user@gmail.com", "password123").Return("accessToken", "refreshToken", nil)

	handler.Login(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.LoginResponse
	err := json.NewDecoder(w.Body).Decode(&resp)

	assert.NoError(t, err)
	assert.Equal(t, "accessToken", resp.AccessToken)
	assert.Equal(t, "refreshToken", resp.RefreshToken)
	mockUC.AssertExpectations(t)
}

func TestHandler_Login_WrongPassword(t *testing.T) {
	mockUC := new(MockUserUseCase)
	handler := NewHandler(mockUC)

	reqBody := `{"email":"user@gmail.com", "password":"wrong_password"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(reqBody))
	w := httptest.NewRecorder()

	mockUC.On("Login", mock.Anything, "user@gmail.com", "wrong_password").Return("", "", domain.ErrInvalidCredentials)

	handler.Login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}