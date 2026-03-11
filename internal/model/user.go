package model

import "time"

type User struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Username  string    `json:"username" gorm:"uniqueIndex;size:64;not null"`
	Password  string    `json:"-" gorm:"size:256;not null"`
	Role      string    `json:"role" gorm:"size:32;not null;default:student"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=admin teacher student"`
}
