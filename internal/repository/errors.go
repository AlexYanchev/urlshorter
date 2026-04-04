package repository

import "errors"

var (
	ErrDublicateID = errors.New("ID already exist")
)