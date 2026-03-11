package repository

import (
	"mcp-agent/internal/constants"
	"mcp-agent/internal/model"
	"time"

	"gorm.io/gorm"
)

type LogRepository struct {
	db *gorm.DB
}

func NewLogRepository(db *gorm.DB) *LogRepository {
	return &LogRepository{db: db}
}

func (r *LogRepository) Create(log *model.CallLog) error {
	return r.db.Create(log).Error
}

func (r *LogRepository) Query(req model.LogQueryRequest) (*model.LogListResponse, error) {
	query := r.db.Model(&model.CallLog{})

	if req.ToolName != "" {
		query = query.Where("tool_name = ?", req.ToolName)
	}
	if req.CallerID > 0 {
		query = query.Where("caller_id = ?", req.CallerID)
	}
	if req.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if req.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
			query = query.Where("created_at <= ?", t)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > constants.MaxPageSize {
		req.PageSize = constants.DefaultPageSize
	}

	var items []model.CallLog
	offset := (req.Page - 1) * req.PageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&items).Error
	if err != nil {
		return nil, err
	}

	return &model.LogListResponse{
		Total: total,
		Items: items,
		Page:  req.Page,
	}, nil
}
