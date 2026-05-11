package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/AlexYanchev/urlshorter/internal/model"
	"github.com/AlexYanchev/urlshorter/internal/repository"
)

type URLRepository interface {
	Save(id, value string) error
	SaveBatch(items []model.BatchURLItem) error
	Get(id string) (string, bool)
	GetByOriginalURL(originalURL string) (string, bool)
	Ping() error
}

type Service struct {
	repository URLRepository
}

type BatchCreateRequest struct {
	CorrelationID string
	OriginalURL   string
}

type BatchCreateResult struct {
	CorrelationID string
	ShortID       string
}

type DuplicateOriginalURLError struct {
	ShortID string
}

func (e *DuplicateOriginalURLError) Error() string {
	return "original URL already exists"
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

		if errors.Is(err, repository.ErrDuplicateID) {
			continue
		}

		var duplicateErr *repository.DuplicateOriginalURLError
		if errors.As(err, &duplicateErr) {
			return "", &DuplicateOriginalURLError{ShortID: duplicateErr.ShortID}
		}

		if errors.Is(err, repository.ErrDuplicateOriginalURL) {
			return "", fmt.Errorf("failed to get existing short url")
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

func (s *Service) CreateShortURLBatch(requests []BatchCreateRequest) ([]BatchCreateResult, error) {
	maxGeneration := 10000

	for range maxGeneration {
		items := make([]model.BatchURLItem, 0, len(requests))
		results := make([]BatchCreateResult, 0, len(requests))
		generatedIDs := make(map[string]struct{}, len(requests))

		for _, request := range requests {
			id, err := generateUniqueBatchID(generatedIDs)
			if err != nil {
				return nil, err
			}

			items = append(items, model.BatchURLItem{
				ShortURL:    id,
				OriginalURL: request.OriginalURL,
			})

			results = append(results, BatchCreateResult{
				CorrelationID: request.CorrelationID,
				ShortID:       id,
			})
		}

		err := s.repository.SaveBatch(items)
		if err == nil {
			return results, nil
		}

		if errors.Is(err, repository.ErrDuplicateID) {
			continue
		}

		return nil, fmt.Errorf("failed to save batch: %w", err)
	}

	return nil, fmt.Errorf("cannot generate unique IDs for batch")
}

func (s *Service) Ping() error {
	return s.repository.Ping()
}

func generateID() string {
	b := make([]byte, 6)

	rand.Read(b)

	return base64.URLEncoding.EncodeToString(b)[:6]
}

func generateUniqueBatchID(generatedIDs map[string]struct{}) (string, error) {
	maxGeneration := 10000

	for range maxGeneration {
		id := generateID()
		if _, exists := generatedIDs[id]; exists {
			continue
		}

		generatedIDs[id] = struct{}{}
		return id, nil
	}

	return "", fmt.Errorf("cannot generate unique ID inside batch")
}
