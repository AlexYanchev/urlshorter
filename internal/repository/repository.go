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
	defer r.mu.Unlock()

	if err := r.saveDataLocked(id, value); err != nil {
		return err
	}

	if r.storagePath != "" {
		if err := r.saveToFileLocked(); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) SaveBatch(items []model.BatchURLItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	batchIDs := make(map[string]struct{}, len(items))
	batchOriginalURLs := make(map[string]struct{}, len(items))

	for _, item := range items {
		if _, exists := batchIDs[item.ShortURL]; exists {
			return ErrDuplicateID
		}
		batchIDs[item.ShortURL] = struct{}{}

		if _, exists := r.data[item.ShortURL]; exists {
			return ErrDuplicateID
		}

		if _, exists := batchOriginalURLs[item.OriginalURL]; exists {
			return ErrDuplicateOriginalURL
		}
		batchOriginalURLs[item.OriginalURL] = struct{}{}

		for _, originalURL := range r.data {
			if originalURL == item.OriginalURL {
				return ErrDuplicateOriginalURL
			}
		}
	}

	for _, item := range items {
		r.data[item.ShortURL] = item.OriginalURL
	}

	if r.storagePath != "" {
		if err := r.saveToFileLocked(); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) saveDataLocked(id, value string) error {
	_, ok := r.data[id]
	if ok {
		return ErrDuplicateID
	}

	for shortURL, originalURL := range r.data {
		if originalURL == value {
			return &DuplicateOriginalURLError{ShortID: shortURL}
		}
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

func (r *Repository) GetByOriginalURL(originalURL string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for shortURL, savedOriginalURL := range r.data {
		if savedOriginalURL == originalURL {
			return shortURL, true
		}
	}

	return "", false
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

func (r *Repository) saveToFileLocked() error {
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
