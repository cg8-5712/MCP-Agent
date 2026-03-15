package repository

import (
	"mcp-agent/internal/model"
	"time"

	"gorm.io/gorm"
)

type StatsRepository struct {
	db *gorm.DB
}

func NewStatsRepository(db *gorm.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

func (r *StatsRepository) GetByToolName(toolName string) (*model.ToolStats, error) {
	var stats model.ToolStats
	err := r.db.Where("tool_name = ?", toolName).First(&stats).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &stats, err
}

func (r *StatsRepository) CreateOrUpdate(stats *model.ToolStats) error {
	var existing model.ToolStats
	err := r.db.Where("tool_name = ?", stats.ToolName).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return r.db.Create(stats).Error
	}
	if err != nil {
		return err
	}

	stats.ID = existing.ID
	return r.db.Save(stats).Error
}

func (r *StatsRepository) IncrementCall(toolName string, success bool, duration int64) error {
	stats, err := r.GetByToolName(toolName)
	if err != nil {
		return err
	}

	if stats == nil {
		stats = &model.ToolStats{
			ToolName: toolName,
		}
	}

	stats.TotalCalls++
	if success {
		stats.SuccessCalls++
	} else {
		stats.FailedCalls++
	}

	totalDuration := stats.AvgDuration * (stats.TotalCalls - 1)
	stats.AvgDuration = (totalDuration + duration) / stats.TotalCalls
	stats.LastCallAt = time.Now()

	return r.CreateOrUpdate(stats)
}

func (r *StatsRepository) UpdateHealthStatus(toolName, status string) error {
	return r.db.Model(&model.ToolStats{}).
		Where("tool_name = ?", toolName).
		Updates(map[string]interface{}{
			"health_status":     status,
			"last_health_check": time.Now(),
		}).Error
}

func (r *StatsRepository) ListAll() ([]model.ToolStats, error) {
	var stats []model.ToolStats
	err := r.db.Find(&stats).Error
	return stats, err
}
