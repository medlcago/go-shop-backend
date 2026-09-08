package repository

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

func HandleError(err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrRecordNotFound
	case err != nil:
		return err
	default:
		return nil
	}
}

func IsRecordNotFound(err error) bool {
	return errors.Is(err, ErrRecordNotFound)
}
