package repository

import "errors"

var (
	ErrDublicateID          = errors.New("ID already exist")
	ErrDuplicateOriginalURL = errors.New("original URL already exists")
)
