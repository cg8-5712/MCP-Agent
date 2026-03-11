package service

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"mcp-agent/internal/config"
	"mcp-agent/internal/model"
	"mcp-agent/internal/repository"
	"mcp-agent/pkg/logger"

	"go.uber.org/zap"
)

type HealthService struct {
	toolRepo   *repository.ToolRepository
	mcpServers []config.MCPServerConfig
	timeout    time.Duration
}

func NewHealthService(toolRepo *repository.ToolRepository, mcpServers []config.MCPServerConfig, timeout time.Duration) *HealthService {
	return &HealthService{
		toolRepo:   toolRepo,
		mcpServers: mcpServers,
		timeout:    timeout,
	}
}

func (s *HealthService) CheckAll() []model.ToolHealthResponse {
	tools, err := s.toolRepo.ListEnabled()
	if err != nil {
		logger.Error("failed to list tools for health check", zap.Error(err))
		return nil
	}

	results := make([]model.ToolHealthResponse, len(tools))
	var wg sync.WaitGroup

	for i, tool := range tools {
		wg.Add(1)
		go func(idx int, t model.Tool) {
			defer wg.Done()
			results[idx] = s.checkTool(t)
		}(i, tool)
	}

	wg.Wait()
	return results
}

func (s *HealthService) CheckTool(name string) (*model.ToolHealthResponse, error) {
	tool, err := s.toolRepo.GetByName(name)
	if err != nil {
		return nil, err
	}
	result := s.checkTool(*tool)
	return &result, nil
}

func (s *HealthService) checkTool(tool model.Tool) model.ToolHealthResponse {
	result := model.ToolHealthResponse{
		Name:   tool.Name,
		Status: "offline",
	}

	url := fmt.Sprintf("%s/health", tool.ServerURL)
	client := &http.Client{Timeout: s.timeout}

	start := time.Now()
	resp, err := client.Get(url)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		logger.Warn("health check failed", zap.String("tool", tool.Name), zap.Error(err))
		result.Latency = latency
		return result
	}
	defer resp.Body.Close()

	result.Latency = latency
	if resp.StatusCode == http.StatusOK {
		result.Status = "online"
	}

	return result
}
