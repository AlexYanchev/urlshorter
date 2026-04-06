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

	_, ok := r.data[id]
	if ok {
		return ErrDublicateID
	}

	r.data[id] = value
	return nil
}

func (r *Repository) Get(id string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	value, ok := r.data[id]

	return value, ok
} 