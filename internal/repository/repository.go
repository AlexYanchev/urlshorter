package repository

import (
	"sync"
)

type Repository struct {
	data map[string]string
	mu   sync.Mutex
}

func New() *Repository {
	return &Repository{
		data: make(map[string]string),
	}
}

func (r *Repository) Save(id, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exist := r.data[id]
	if exist {
		return ErrDublicateID
	}

	r.data[id] = value
	return nil
}

func (r *Repository) Get(id string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	value, exists := r.data[id]

	return value, exists
} 