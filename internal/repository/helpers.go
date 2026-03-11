package repository

import (
	"errors"

	"gorm.io/gorm"
)

func HandleNotFoundError(err error, notFoundErr error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return notFoundErr
	}
	return err
}
