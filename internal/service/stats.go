package service

import (
	"mcp-agent/internal/model"
	"mcp-agent/internal/repository"
)

type StatsService struct {
	statsRepo *repository.StatsRepository
}

func NewStatsService(statsRepo *repository.StatsRepository) *StatsService {
	return &StatsService{statsRepo: statsRepo}
}

func (s *StatsService) GetToolStats(toolName string) (*model.ToolStatsResponse, error) {
	stats, err := s.statsRepo.GetByToolName(toolName)
	if err != nil {
		return nil, err
	}
	if stats == nil {
		return &model.ToolStatsResponse{
			ToolName:     toolName,
			HealthStatus: "unknown",
		}, nil
	}

	successRate := 0.0
	if stats.TotalCalls > 0 {
		successRate = float64(stats.SuccessCalls) / float64(stats.TotalCalls) * 100
	}

	return &model.ToolStatsResponse{
		ToolName:     stats.ToolName,
		TotalCalls:   stats.TotalCalls,
		SuccessCalls: stats.SuccessCalls,
		FailedCalls:  stats.FailedCalls,
		SuccessRate:  successRate,
		AvgDuration:  stats.AvgDuration,
		HealthStatus: stats.HealthStatus,
		LastCallAt:   stats.LastCallAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *StatsService) ListAllStats() ([]model.ToolStatsResponse, error) {
	statsList, err := s.statsRepo.ListAll()
	if err != nil {
		return nil, err
	}

	result := make([]model.ToolStatsResponse, len(statsList))
	for i, stats := range statsList {
		successRate := 0.0
		if stats.TotalCalls > 0 {
			successRate = float64(stats.SuccessCalls) / float64(stats.TotalCalls) * 100
		}

		result[i] = model.ToolStatsResponse{
			ToolName:     stats.ToolName,
			TotalCalls:   stats.TotalCalls,
			SuccessCalls: stats.SuccessCalls,
			FailedCalls:  stats.FailedCalls,
			SuccessRate:  successRate,
			AvgDuration:  stats.AvgDuration,
			HealthStatus: stats.HealthStatus,
			LastCallAt:   stats.LastCallAt.Format("2006-01-02 15:04:05"),
		}
	}

	return result, nil
}
