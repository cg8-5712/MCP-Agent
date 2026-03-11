package repository

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrToolNotFound = errors.New("tool not found")
	ErrLogNotFound  = errors.New("log not found")
	ErrDuplicateKey = errors.New("duplicate key")
)
