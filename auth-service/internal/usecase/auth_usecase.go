package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/MiRRoRise/auth-service/internal/domain"
	"github.com/MiRRoRise/auth-service/internal/kafka"
	"github.com/MiRRoRise/auth-service/internal/metrics"
	"github.com/MiRRoRise/auth-service/internal/repository"
	"github.com/MiRRoRise/auth-service/pkg/jwt"
	"github.com/MiRRoRise/auth-service/pkg/logger"
	"github.com/MiRRoRise/auth-service/pkg/password"
)

type UserUseCase interface {
	Register(ctx context.Context, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	RefreshTokens(ctx context.Context, refreshToken string) (newAccess, newRefresh string, err error)
	GetUserByID(ctx context.Context, userID int64) (*domain.User, error)
}

type userUseCase struct {
	repo      repository.UserRepository
	hasher    password.Hasher
	tokens    jwt.TokenManager
	publisher kafka.EventPublisher
	logger    *logger.Logger
}

func NewUserUseCase(
	repo repository.UserRepository,
	hasher password.Hasher,
	tokens jwt.TokenManager,
	publisher kafka.EventPublisher,
	logger *logger.Logger,
) *userUseCase {
	if publisher == nil {
		publisher = kafka.NoopPublisher{}
	}
	return &userUseCase{
		repo:      repo,
		hasher:    hasher,
		tokens:    tokens,
		publisher: publisher,
		logger:    logger,
	}
}

var _ UserUseCase = (*userUseCase)(nil)

func (u *userUseCase) Register(ctx context.Context, email, password string) (*domain.User, error) {
	existing, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user: %w", err)
	}
	if existing != nil {
		return nil, domain.ErrUserExists
	}

	passwordHash, err := u.hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: passwordHash,
	}

	err = u.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("sql error create user: %w", err)
	}

	metrics.RegistrationsTotal.Inc()

	event := kafka.UserRegisteredEvent{
		UserID:    user.ID,
		Email:     user.Email,
		CreatedAt: time.Now().UTC(),
	}
	if err := u.publisher.PublishUserRegistered(event); err != nil {
		if u.logger != nil {
			u.logger.Error("failed to publish user.registered", err)
		}
	}

	return user, nil
}

func (u *userUseCase) Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error) {
	existing, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", fmt.Errorf("failed to check user: %w", err)
	}
	if existing == nil {
		return "", "", domain.ErrUserNotFound
	}

	if compareErr := u.hasher.Compare(existing.PasswordHash, password); compareErr != nil {
		return "", "", domain.ErrInvalidCredentials
	}

	accessToken, err = u.tokens.GenerateAccessToken(existing.ID)
	if err != nil {
		return "", "", fmt.Errorf("error generate access token: %w", err)
	}

	refreshToken, err = u.tokens.GenerateRefreshToken(existing.ID)
	if err != nil {
		return "", "", fmt.Errorf("error generate refresh token: %w", err)
	}

	metrics.LoginsTotal.Inc()

	return accessToken, refreshToken, nil
}

func (u *userUseCase) RefreshTokens(ctx context.Context, refreshToken string) (newAccess, newRefresh string, err error) {
	userID, err := u.tokens.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", domain.ErrInvalidToken
	}

	user, err := u.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return "", "", domain.ErrUserNotFound
	}

	newAccess, err = u.tokens.GenerateAccessToken(userID)
	if err != nil {
		return "", "", fmt.Errorf("error generate access token: %w", err)
	}

	newRefresh, err = u.tokens.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", fmt.Errorf("error generate refresh token: %w", err)
	}

	return newAccess, newRefresh, nil
}

func (u *userUseCase) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	user, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}
