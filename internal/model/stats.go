package model

import "time"

type ToolStats struct {
	ID              int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ToolName        string    `json:"tool_name" gorm:"uniqueIndex;size:128;not null"`
	TotalCalls      int64     `json:"total_calls" gorm:"default:0"`
	SuccessCalls    int64     `json:"success_calls" gorm:"default:0"`
	FailedCalls     int64     `json:"failed_calls" gorm:"default:0"`
	AvgDuration     int64     `json:"avg_duration_ms" gorm:"default:0"`
	LastCallAt      time.Time `json:"last_call_at"`
	LastHealthCheck time.Time `json:"last_health_check"`
	HealthStatus    string    `json:"health_status" gorm:"size:32;default:unknown"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ToolStatsResponse struct {
	ToolName     string  `json:"tool_name"`
	TotalCalls   int64   `json:"total_calls"`
	SuccessCalls int64   `json:"success_calls"`
	FailedCalls  int64   `json:"failed_calls"`
	SuccessRate  float64 `json:"success_rate"`
	AvgDuration  int64   `json:"avg_duration_ms"`
	HealthStatus string  `json:"health_status"`
	LastCallAt   string  `json:"last_call_at"`
}
