package service

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
)

type Service struct {
	storage map[string]string
	mu      sync.RWMutex
}

func New() *Service {
	return &Service{
		storage: make(map[string]string),
	}
}

func (s *Service) CreateShortURL(originalURL string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := generateID()

	s.storage[id] = originalURL

	return id
}

func (s *Service) GetOriginalURL(id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	originalURL, exists := s.storage[id]

	return originalURL, exists
}

func generateID() string {
	b := make([]byte, 6)

	rand.Read(b)
	
	return base64.URLEncoding.EncodeToString(b)[:6]
}