package usecase

import (
	"context"
	"fmt"

	"github.com/MiRRoRise/auth-service/internal/domain"
	"github.com/MiRRoRise/auth-service/internal/repository"
	"github.com/MiRRoRise/auth-service/pkg/jwt"
	"github.com/MiRRoRise/auth-service/pkg/password"
)

type UserUseCase interface {
	Register(ctx context.Context, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	RefreshTokens(ctx context.Context, refreshToken string) (newAccess, newRefresh string, err error)
	GetUserByID(ctx context.Context, userID int64) (*domain.User, error)
	ValidateRefreshToken(ctx context.Context, tokenString string) (int64, error)
}

type userUseCase struct {
	repo   repository.UserRepository
	hasher password.Hasher
	tokens jwt.TokenManager
}

func NewUserUseCase(
	repo repository.UserRepository,
	hasher password.Hasher,
	tokens jwt.TokenManager,
) *userUseCase {
	return &userUseCase{
		repo:   repo,
		hasher: hasher,
		tokens: tokens,
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

	if err := u.hasher.Compare(existing.PasswordHash, password); err != nil {
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

func (u *userUseCase) ValidateRefreshToken(ctx context.Context, tokenString string) (int64, error) {
	return u.tokens.ValidateRefreshToken(tokenString)
}
