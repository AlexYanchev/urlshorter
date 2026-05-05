package repository

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/AlexYanchev/urlshorter/internal/model"
)

type Repository struct {
	data        map[string]string
	mu          sync.Mutex
	storagePath string
}

func New(storagePath string) *Repository {
	repo := &Repository{
		data:        make(map[string]string),
		storagePath: storagePath,
	}

	if storagePath != "" {
		if err := repo.load(); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: failed to load from file: %v", err)
		}
	}

	return repo
}

func (r *Repository) Save(id, value string) error {
	r.mu.Lock()

	_, ok := r.data[id]
	if ok {
		return ErrDublicateID
	}

	r.data[id] = value

	r.mu.Unlock()

	if r.storagePath != "" {
		if err := r.saveToFile(); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) Get(id string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	value, ok := r.data[id]

	return value, ok
}

func (r *Repository) Ping() error {
	return fmt.Errorf("database is not configured")
}

func (r *Repository) load() error {
	file, err := os.Open(r.storagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var items []model.URLItem
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&items); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.data = make(map[string]string, len(items))
	for _, item := range items {
		r.data[item.ShortURL] = item.OriginalURL
	}

	return nil
}

func (r *Repository) saveToFile() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	items := make([]model.URLItem, 0, len(r.data))
	for shortURL, originalURL := range r.data {
		items = append(items, model.URLItem{
			UUID:        generateUUID(),
			ShortURL:    shortURL,
			OriginalURL: originalURL,
		})
	}

	file, err := os.Create(r.storagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(items)
}
