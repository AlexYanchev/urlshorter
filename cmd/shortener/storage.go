package main

import (
	"fmt"

	"github.com/AlexYanchev/urlshorter/internal/config"
	"github.com/AlexYanchev/urlshorter/internal/migrations"
	"github.com/AlexYanchev/urlshorter/internal/repository"
	"github.com/AlexYanchev/urlshorter/internal/service"
)

type closeFunc func() error

func buildRepository(cfg *config.Config) (service.URLRepository, closeFunc, error) {
	if cfg.DatabaseDSN != "" {
		postgresRepo, err := repository.NewPostgres(cfg.DatabaseDSN)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to connect database: %w", err)
		}

		if err := migrations.Run(postgresRepo.DB()); err != nil {
			postgresRepo.Close()
			return nil, nil, fmt.Errorf("failed to apply migrations: %w", err)
		}

		return postgresRepo, postgresRepo.Close, nil
	}

	if cfg.FileStoragePath != "" {
		return repository.New(cfg.FileStoragePath), nil, nil
	}

	return repository.New(""), nil, nil
}
