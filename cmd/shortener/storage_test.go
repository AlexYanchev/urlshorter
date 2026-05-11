package main

import (
	"testing"

	"github.com/AlexYanchev/urlshorter/internal/config"
	"github.com/AlexYanchev/urlshorter/internal/repository"
)

func TestBuildRepository_UsesFileStorageWhenPathProvided(t *testing.T) {
	cfg := &config.Config{
		FileStoragePath: "test-storage.json",
	}

	repo, closeFn, err := buildRepository(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if closeFn != nil {
		t.Fatal("expected close function to be nil for file storage")
	}

	fileRepo, ok := repo.(*repository.Repository)
	if !ok {
		t.Fatalf("expected *repository.Repository, got %T", repo)
	}

	if fileRepo == nil {
		t.Fatal("expected non nil repository")
	}
}

func TestBuildRepository_UsesMemoryStorageWhenPathIsEmpty(t *testing.T) {
	cfg := &config.Config{}

	repo, closeFn, err := buildRepository(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if closeFn != nil {
		t.Fatal("expected close function to be nil for memory storage")
	}

	memoryRepo, ok := repo.(*repository.Repository)
	if !ok {
		t.Fatalf("expected *repository.Repository, got %T", repo)
	}

	if memoryRepo == nil {
		t.Fatal("expected non nil repository")
	}
}
