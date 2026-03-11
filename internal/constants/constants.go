package constants

import "time"

const (
	AccessTokenExpiry  = 24 * time.Hour
	RefreshTokenExpiry = 7 * 24 * time.Hour

	DefaultPageSize = 20
	MaxPageSize     = 100

	ToolCallTimeout = 30 * time.Second

	HealthCheckInterval = 30 * time.Second
	HealthCheckTimeout  = 3 * time.Second

	RoleAdmin   = "admin"
	RoleTeacher = "teacher"
	RoleStudent = "student"
)
