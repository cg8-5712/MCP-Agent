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
	"mcp-agent/internal/repository"
	"mcp-agent/pkg/logger"

	"go.uber.org/zap"
)

type ToolService struct {
	toolRepo *repository.ToolRepository
	logRepo  *repository.LogRepository
}

func NewToolService(toolRepo *repository.ToolRepository, logRepo *repository.LogRepository) *ToolService {
	return &ToolService{
		toolRepo: toolRepo,
		logRepo:  logRepo,
	}
}

func (s *ToolService) Create(req model.CreateToolRequest) (*model.Tool, error) {
	tool := &model.Tool{
		Name:        req.Name,
		Description: req.Description,
		ServerName:  req.ServerName,
		ServerURL:   req.ServerURL,
		Schema:      req.Schema,
		Enabled:     true,
		Version:     req.Version,
	}
	if tool.Version == "" {
		tool.Version = "1.0.0"
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

	if callErr != nil {
		return nil, fmt.Errorf("%w: %s", ErrToolCallFailed, callErr.Error())
	}
	return result, nil
}

func (s *ToolService) doCall(tool *model.Tool, args map[string]interface{}) (interface{}, error) {
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
