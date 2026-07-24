package domain

import "errors"

var (
	ErrChatNotFound     = errors.New("chat not found")
	ErrMessageNotFound  = errors.New("message not found")
	ErrChatExists       = errors.New("chat already exist")
	ErrInvalidMessage   = errors.New("invalid message")
	ErrChatNameRequired = errors.New("chat name is required")
	ErrMessageEmpty     = errors.New("message cannot be empty")
	ErrChatIDRequired   = errors.New("chat id reqired")
	ErrUnauthorized     = errors.New("user unauthorized")
	ErrInvalidToken     = errors.New("invalid token")
	ErrUserNotFound     = errors.New("user not found")
)
