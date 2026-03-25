package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"mcp-agent/internal/model"
	"mcp-agent/internal/native"
	"mcp-agent/internal/repository"
	"mcp-agent/pkg/logger"

	"go.uber.org/zap"
)

type ToolService struct {
	toolRepo      *repository.ToolRepository
	logRepo       *repository.LogRepository
	statsRepo     *repository.StatsRepository
	nativeExec    *native.Executor
}

func NewToolService(toolRepo *repository.ToolRepository, logRepo *repository.LogRepository, statsRepo *repository.StatsRepository) *ToolService {
	return &ToolService{
		toolRepo:   toolRepo,
		logRepo:    logRepo,
		statsRepo:  statsRepo,
		nativeExec: native.NewExecutor(),
	}
}

func (s *ToolService) Create(req model.CreateToolRequest) (*model.Tool, error) {
	tool := &model.Tool{
		Name:         req.Name,
		Description:  req.Description,
		ServerName:   req.ServerName,
		ServerURL:    req.ServerURL,
		Schema:       req.Schema,
		Enabled:      true,
		Version:      req.Version,
		ToolType:     req.ToolType,
		NativeConfig: req.NativeConfig,
	}
	if tool.Version == "" {
		tool.Version = "1.0.0"
	}
	if tool.ToolType == "" {
		tool.ToolType = model.ToolTypeRemote
	}
	if err := s.toolRepo.Create(tool); err != nil {
		return nil, err
	}
	return tool, nil
}

func (s *ToolService) GetByName(name string) (*model.Tool, error) {
	tool, err := s.toolRepo.GetByName(name)
	if err != nil {
		if errors.Is(err, repository.ErrToolNotFound) {
			return nil, ErrToolNotFound
		}
		return nil, err
	}
	return tool, nil
}

func (s *ToolService) List() ([]model.Tool, error) {
	return s.toolRepo.List()
}

func (s *ToolService) ListEnabled() ([]model.Tool, error) {
	return s.toolRepo.ListEnabled()
}

func (s *ToolService) Update(name string, req model.UpdateToolRequest) (*model.Tool, error) {
	tool, err := s.toolRepo.GetByName(name)
	if err != nil {
		if errors.Is(err, repository.ErrToolNotFound) {
			return nil, ErrToolNotFound
		}
		return nil, err
	}

	if req.Description != nil {
		tool.Description = *req.Description
	}
	if req.ServerURL != nil {
		tool.ServerURL = *req.ServerURL
	}
	if req.Schema != nil {
		tool.Schema = req.Schema
	}
	if req.Enabled != nil {
		tool.Enabled = *req.Enabled
	}
	if req.Version != nil {
		tool.Version = *req.Version
	}

	if err := s.toolRepo.Update(tool); err != nil {
		return nil, err
	}
	return tool, nil
}

func (s *ToolService) Delete(name string) error {
	return s.toolRepo.Delete(name)
}

func (s *ToolService) CallTool(name string, args map[string]interface{}, callerID int64, callerName string) (interface{}, error) {
	tool, err := s.toolRepo.GetByName(name)
	if err != nil {
		if errors.Is(err, repository.ErrToolNotFound) {
			return nil, ErrToolNotFound
		}
		return nil, err
	}

	if !tool.Enabled {
		return nil, ErrToolDisabled
	}

	// Schema 校验
	if tool.Schema != nil && len(tool.Schema) > 0 {
		if err := s.validateSchema(tool.Schema, args); err != nil {
			return nil, fmt.Errorf("schema validation failed: %w", err)
		}
	}

	start := time.Now()

	callLog := &model.CallLog{
		ToolName:   name,
		CallerID:   callerID,
		CallerName: callerName,
	}

	inputBytes, _ := json.Marshal(args)
	callLog.Input = string(inputBytes)

	result, callErr := s.doCall(tool, args)
	duration := time.Since(start).Milliseconds()

	callLog.Duration = duration
	success := callErr == nil
	if callErr != nil {
		callLog.StatusCode = 500
		callLog.Error = callErr.Error()
	} else {
		callLog.StatusCode = 200
		outputBytes, _ := json.Marshal(result)
		callLog.Output = string(outputBytes)
	}

	if logErr := s.logRepo.Create(callLog); logErr != nil {
		logger.Error("failed to save call log", zap.Error(logErr))
	}

	// 更新统计信息
	if s.statsRepo != nil {
		if err := s.statsRepo.IncrementCall(name, success, duration); err != nil {
			logger.Error("failed to update stats", zap.Error(err))
		}
	}

	if callErr != nil {
		return nil, fmt.Errorf("%w: %s", ErrToolCallFailed, callErr.Error())
	}
	return result, nil
}

func (s *ToolService) validateSchema(schema map[string]interface{}, args map[string]interface{}) error {
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return nil
	}

	required, _ := schema["required"].([]interface{})
	requiredFields := make(map[string]bool)
	for _, field := range required {
		if fieldName, ok := field.(string); ok {
			requiredFields[fieldName] = true
		}
	}

	for field := range requiredFields {
		if _, exists := args[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	for field, value := range args {
		propSchema, exists := properties[field]
		if !exists {
			continue
		}

		propMap, ok := propSchema.(map[string]interface{})
		if !ok {
			continue
		}

		expectedType, ok := propMap["type"].(string)
		if !ok {
			continue
		}

		actualType := getValueType(value)
		if actualType != expectedType && expectedType != "any" {
			return fmt.Errorf("field %s: expected type %s, got %s", field, expectedType, actualType)
		}
	}

	return nil
}

func getValueType(value interface{}) string {
	if value == nil {
		return "null"
	}
	switch value.(type) {
	case string:
		return "string"
	case float64, int, int64:
		return "number"
	case bool:
		return "boolean"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}

func (s *ToolService) doCall(tool *model.Tool, args map[string]interface{}) (interface{}, error) {
	switch tool.ToolType {
	case model.ToolTypeNative:
		return s.nativeExec.Execute(tool, args)
	case model.ToolTypeCustom:
		if tool.ServerURL == "" {
			return nil, fmt.Errorf("custom tool %q has no server_url", tool.Name)
		}
		return s.doRemoteCall(tool, args)
	default: // ToolTypeRemote or empty
		return s.doRemoteCall(tool, args)
	}
}

func (s *ToolService) doRemoteCall(tool *model.Tool, args map[string]interface{}) (interface{}, error) {
	body, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/call", tool.ServerURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tool server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return string(respBody), nil
	}
	return result, nil
}
