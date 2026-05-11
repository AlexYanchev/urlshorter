package repository

import "errors"

var (
	ErrDuplicateID          = errors.New("ID already exist")
	ErrDuplicateOriginalURL = errors.New("original URL already exists")
)

type DuplicateOriginalURLError struct {
	ShortID string
}

func (e *DuplicateOriginalURLError) Error() string {
	return "original URL already exists"
}
