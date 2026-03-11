package repository

import (
	"mcp-agent/internal/model"

	"gorm.io/gorm"
)

type ToolRepository struct {
	db *gorm.DB
}

func NewToolRepository(db *gorm.DB) *ToolRepository {
	return &ToolRepository{db: db}
}

func (r *ToolRepository) Create(tool *model.Tool) error {
	return r.db.Create(tool).Error
}

func (r *ToolRepository) GetByName(name string) (*model.Tool, error) {
	var tool model.Tool
	err := r.db.Where("name = ?", name).First(&tool).Error
	return &tool, HandleNotFoundError(err, ErrToolNotFound)
}

func (r *ToolRepository) GetByID(id int64) (*model.Tool, error) {
	var tool model.Tool
	err := r.db.First(&tool, id).Error
	return &tool, HandleNotFoundError(err, ErrToolNotFound)
}

func (r *ToolRepository) List() ([]model.Tool, error) {
	var tools []model.Tool
	err := r.db.Find(&tools).Error
	return tools, err
}

func (r *ToolRepository) ListEnabled() ([]model.Tool, error) {
	var tools []model.Tool
	err := r.db.Where("enabled = ?", true).Find(&tools).Error
	return tools, err
}

func (r *ToolRepository) Update(tool *model.Tool) error {
	return r.db.Save(tool).Error
}

func (r *ToolRepository) Delete(name string) error {
	result := r.db.Where("name = ?", name).Delete(&model.Tool{})
	if result.RowsAffected == 0 {
		return ErrToolNotFound
	}
	return result.Error
}
