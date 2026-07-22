package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exist")
	ErrInvalidCredentials = errors.New("incorrect data")
	ErrUnauthorized       = errors.New("user unauthorized")
	ErrInvalidToken       = errors.New("invalid token")
)
