package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}",  nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		s, ok := value.(string)
		if !ok {
			return errors.New("failed to scan JSONMap")
		}
		bytes = []byte(s)
	}
	return json.Unmarshal(bytes, j)
}

// ToolType defines how the tool is executed.
// - "remote": forward call to an external MCP Server (original behaviour)
// - "native": executed directly by MCP-Agent using a built-in template
// - "custom": user-supplied endpoint + schema, called via plain HTTP
type ToolType string

const (
	ToolTypeRemote ToolType = "remote"
	ToolTypeNative ToolType = "native"
	ToolTypeCustom ToolType = "custom"
)

// NativeTemplate is the built-in execution template for native tools.
type NativeTemplate string

const (
	NativeTplHTTPGet  NativeTemplate = "http_get"
	NativeTplHTTPPost NativeTemplate = "http_post"
	NativeTplReadDoc  NativeTemplate = "read_doc"
)

type Tool struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name         string    `json:"name" gorm:"uniqueIndex;size:128;not null"`
	Description  string    `json:"description" gorm:"size:512"`
	ToolType     ToolType  `json:"tool_type" gorm:"size:32;default:remote"`
	ServerName   string    `json:"server_name" gorm:"size:128"`
	ServerURL    string    `json:"server_url" gorm:"size:256"`
	NativeConfig JSONMap   `json:"native_config" gorm:"type:text"`
	Schema       JSONMap   `json:"schema" gorm:"type:text"`
	Enabled      bool      `json:"enabled" gorm:"default:true"`
	Version      string    `json:"version" gorm:"size:32;default:1.0.0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateToolRequest struct {
	Name         string   `json:"name" binding:"required"`
	Description  string   `json:"description"`
	ToolType     ToolType `json:"tool_type"`       // remote | native | custom; default remote
	ServerName   string   `json:"server_name"`
	ServerURL    string   `json:"server_url"`      // required for remote & custom
	NativeConfig JSONMap  `json:"native_config"`   // required for native
	Schema       JSONMap  `json:"schema"`
	Version      string   `json:"version"`
}

type UpdateToolRequest struct {
	Description  *string  `json:"description"`
	ServerURL    *string  `json:"server_url"`
	NativeConfig JSONMap  `json:"native_config"`
	Schema       JSONMap  `json:"schema"`
	Enabled      *bool    `json:"enabled"`
	Version      *string  `json:"version"`
}

type CallToolRequest struct {
	Arguments map[string]interface{} `json:"arguments"`
}

type ToolHealthResponse struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Latency int64  `json:"latency_ms"`
}
