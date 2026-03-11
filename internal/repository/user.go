package repository

import (
	"mcp-agent/internal/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetByID(id int64) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	return &user, HandleNotFoundError(err, ErrUserNotFound)
}

func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, HandleNotFoundError(err, ErrUserNotFound)
}

func (r *UserRepository) List() ([]model.User, error) {
	var users []model.User
	err := r.db.Find(&users).Error
	return users, err
}
