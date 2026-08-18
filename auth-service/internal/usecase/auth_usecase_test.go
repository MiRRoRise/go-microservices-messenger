package usecase

import (
	"context"
	"testing"

	"github.com/MiRRoRise/auth-service/internal/domain"
	"github.com/MiRRoRise/auth-service/pkg/password"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

type MockManager struct {
	mock.Mock
}

func (m *MockManager) GenerateAccessToken(id int64) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func (m *MockManager) GenerateRefreshToken(id int64) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func (m *MockManager) ValidateAccessToken(tokenString string) (int64, error) {
	args := m.Called(tokenString)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockManager) ValidateRefreshToken(tokenString string) (int64, error) {
	args := m.Called(tokenString)
	return args.Get(0).(int64), args.Error(1)
}

func TestAuthUsecase_Register_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockManager := new(MockManager)
	hasher := password.NewBcryptHasher(4)

	usecase := NewUserUseCase(mockRepo, hasher, mockManager, nil, nil)
	ctx := context.Background()

	mockRepo.On("GetByEmail", ctx, "correct@email.com").Return(nil, nil)
	mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	user, err := usecase.Register(ctx, "correct@email.com", "password123")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "correct@email.com", user.Email)
	assert.NotEmpty(t, user.PasswordHash)
	mockRepo.AssertExpectations(t)
}

func TestAuthUsecase_Register_UserExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockManager := new(MockManager)
	hasher := password.NewBcryptHasher(4)

	usecase := NewUserUseCase(mockRepo, hasher, mockManager, nil, nil)
	ctx := context.Background()

	existing := &domain.User{
		Email:        "userExists@gmail.com",
		PasswordHash: "hashed_password",
	}

	mockRepo.On("GetByEmail", ctx, "userExists@gmail.com").Return(existing, nil)

	user, err := usecase.Register(ctx, "userExists@gmail.com", "password123")

	assert.Error(t, err)
	assert.EqualError(t, err, domain.ErrUserExists.Error())
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestAuthUsecase_Login_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockManager := new(MockManager)
	hasher := password.NewBcryptHasher(4)

	usecase := NewUserUseCase(mockRepo, hasher, mockManager, nil, nil)
	ctx := context.Background()

	hashed, _ := hasher.Hash("password123")
	existing := &domain.User{
		ID:           int64(123),
		Email:        "userExists@gmail.com",
		PasswordHash: hashed,
	}

	mockRepo.On("GetByEmail", ctx, "userExists@gmail.com").Return(existing, nil)
	mockManager.On("GenerateAccessToken", existing.ID).Return("accessToken", nil)
	mockManager.On("GenerateRefreshToken", existing.ID).Return("refreshToken", nil)

	accessToken, refreshToken, err := usecase.Login(ctx, "userExists@gmail.com", "password123")

	assert.NoError(t, err)
	assert.Equal(t, "accessToken", accessToken)
	assert.Equal(t, "refreshToken", refreshToken)
	mockRepo.AssertExpectations(t)
	mockManager.AssertExpectations(t)
}

func TestAuthUsecase_Login_WrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockManager := new(MockManager)
	hasher := password.NewBcryptHasher(4)

	usecase := NewUserUseCase(mockRepo, hasher, mockManager, nil, nil)
	ctx := context.Background()

	hashed, _ := hasher.Hash("correctPassword")
	existing := &domain.User{
		Email:        "userExists@gmail.com",
		PasswordHash: hashed,
	}

	mockRepo.On("GetByEmail", ctx, "userExists@gmail.com").Return(existing, nil)

	accessToken, refreshToken, err := usecase.Login(ctx, "userExists@gmail.com", "incorrectPassword")

	assert.Error(t, err)
	assert.EqualError(t, err, domain.ErrInvalidCredentials.Error())
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	mockRepo.AssertExpectations(t)
}

func TestAuthUsecase_Login_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockManager := new(MockManager)
	hasher := password.NewBcryptHasher(4)

	usecase := NewUserUseCase(mockRepo, hasher, mockManager, nil, nil)
	ctx := context.Background()

	mockRepo.On("GetByEmail", ctx, "userNotExists@gmail.com").Return(nil, nil)

	accessToken, refreshToken, err := usecase.Login(ctx, "userNotExists@gmail.com", "password123")

	assert.Error(t, err)
	assert.EqualError(t, err, domain.ErrUserNotFound.Error())
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	mockRepo.AssertExpectations(t)
}

func TestAuthUsecase_RefreshTokens_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockManager := new(MockManager)
	hasher := password.NewBcryptHasher(4)

	usecase := NewUserUseCase(mockRepo, hasher, mockManager, nil, nil)
	ctx := context.Background()

	user := &domain.User{
		ID:           int64(123),
		Email:        "user@gmail.com",
		PasswordHash: "hashed_password",
	}

	mockManager.On("ValidateRefreshToken", "valid_token").Return(user.ID, nil)
	mockRepo.On("GetByID", ctx, user.ID).Return(user, nil)
	mockManager.On("GenerateAccessToken", user.ID).Return("accessToken", nil)
	mockManager.On("GenerateRefreshToken", user.ID).Return("refreshToken", nil)

	accessToken, refreshToken, err := usecase.RefreshTokens(ctx, "valid_token")

	assert.NoError(t, err)
	assert.Equal(t, "accessToken", accessToken)
	assert.Equal(t, "refreshToken", refreshToken)
	mockManager.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestAuthUsecase_RefreshTokens_InvalidToken(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockManager := new(MockManager)
	hasher := password.NewBcryptHasher(4)

	usecase := NewUserUseCase(mockRepo, hasher, mockManager, nil, nil)
	ctx := context.Background()

	mockManager.On("ValidateRefreshToken", "invalid_token").Return(int64(0), domain.ErrInvalidToken)

	accessToken, refreshToken, err := usecase.RefreshTokens(ctx, "invalid_token")

	assert.Error(t, err)
	assert.EqualError(t, err, domain.ErrInvalidToken.Error())
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	mockManager.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}
