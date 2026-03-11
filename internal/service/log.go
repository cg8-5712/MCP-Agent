package service

import (
	"mcp-agent/internal/model"
	"mcp-agent/internal/repository"
)

type LogService struct {
	logRepo *repository.LogRepository
}

func NewLogService(logRepo *repository.LogRepository) *LogService {
	return &LogService{logRepo: logRepo}
}

func (s *LogService) Query(req model.LogQueryRequest) (*model.LogListResponse, error) {
	return s.logRepo.Query(req)
}
