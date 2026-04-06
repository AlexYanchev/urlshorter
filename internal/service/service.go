package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/AlexYanchev/urlshorter/internal/repository"
)

type URLRepository interface {
	Save(id, value string) error
	Get(id string) (string, bool)
}

type Service struct {
	repository URLRepository
}

func New(r URLRepository) *Service {
	return &Service{
		repository: r,
	}
}

func (s *Service) CreateShortURL(originalURL string) (string, error) {
	maxGeneration := 10000

	for range maxGeneration {
		id := generateID()
		err := s.repository.Save(id, originalURL)

		if err == nil {
			return id, nil
		}

		if errors.Is(err, repository.ErrDublicateID) {
			continue
		}

		return "", fmt.Errorf("failed to save: %w", err)
	}

	return "", fmt.Errorf("cannot generate unique ID")
}

func (s *Service) GetOriginalURL(id string) (string, error) {
	originalURL, exists := s.repository.Get(id)
	if !exists {
		return "", fmt.Errorf("url not exist")
	}

	return originalURL, nil
}

func generateID() string {
	b := make([]byte, 6)

	rand.Read(b)
	
	return base64.URLEncoding.EncodeToString(b)[:6]
}