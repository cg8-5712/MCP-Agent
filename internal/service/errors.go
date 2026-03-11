package service

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrToolNotFound       = errors.New("tool not found")
	ErrToolDisabled       = errors.New("tool is disabled")
	ErrToolCallFailed     = errors.New("tool call failed")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrUserExists         = errors.New("user already exists")
)
