package service

import (
	"errors"
	"testing"

	"github.com/AlexYanchev/urlshorter/internal/model"
	"github.com/AlexYanchev/urlshorter/internal/repository"
)

type batchRepoMock struct {
	saveBatchCalls int
	saveBatchErr   error
}

func (m *batchRepoMock) Save(id, value string) error {
	return nil
}

func (m *batchRepoMock) SaveBatch(items []model.BatchURLItem) error {
	m.saveBatchCalls++
	return m.saveBatchErr
}

func (m *batchRepoMock) Get(id string) (string, bool) {
	return "", false
}

func (m *batchRepoMock) GetByOriginalURL(originalURL string) (string, bool) {
	return "", false
}

func (m *batchRepoMock) Ping() error {
	return nil
}

func TestService_CreateShortURLBatch_DoesNotRetryOnDuplicateOriginalURL(t *testing.T) {
	repoMock := &batchRepoMock{
		saveBatchErr: repository.ErrDuplicateOriginalURL,
	}

	svc := New(repoMock)

	_, err := svc.CreateShortURLBatch([]BatchCreateRequest{
		{CorrelationID: "1", OriginalURL: "http://example.com"},
	})
	if !errors.Is(err, repository.ErrDuplicateOriginalURL) {
		t.Fatalf("expected duplicate original url error, got %v", err)
	}

	if repoMock.saveBatchCalls != 1 {
		t.Fatalf("expected exactly one SaveBatch call, got %d", repoMock.saveBatchCalls)
	}
}
