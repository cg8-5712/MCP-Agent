package model

import "time"

type CallLog struct {
	ID         int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ToolName   string    `json:"tool_name" gorm:"index;size:128;not null"`
	CallerID   int64     `json:"caller_id" gorm:"index"`
	CallerName string    `json:"caller_name" gorm:"size:64"`
	Input      string    `json:"input" gorm:"type:text"`
	Output     string    `json:"output" gorm:"type:text"`
	StatusCode int       `json:"status_code"`
	Duration   int64     `json:"duration_ms"`
	Error      string    `json:"error" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at" gorm:"index"`
}

type LogQueryRequest struct {
	ToolName  string `form:"tool_name"`
	CallerID  int64  `form:"caller_id"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"page_size,default=20"`
}

type LogListResponse struct {
	Total int64      `json:"total"`
	Items []CallLog  `json:"items"`
	Page  int        `json:"page"`
}
