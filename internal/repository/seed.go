package repository

import (
	"mcp-agent/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		return nil
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := model.User{
		Username: "admin",
		Password: string(hashed),
		Role:     "admin",
	}
	return db.Create(&admin).Error
}
