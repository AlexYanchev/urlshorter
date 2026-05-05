package main

import (
	"log"
	"net/http"

	"github.com/AlexYanchev/urlshorter/internal/config"
	"github.com/AlexYanchev/urlshorter/internal/handler"
	"github.com/AlexYanchev/urlshorter/internal/logger"
	"github.com/AlexYanchev/urlshorter/internal/middlewares"
	"github.com/AlexYanchev/urlshorter/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	err := logger.InitLogger()
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}

	defer logger.SugaredLogger.Sync()

	config := config.NewConfig()
	repo, closeRepo, err := buildRepository(config)
	if err != nil {
		logger.SugaredLogger.Fatalw("failed to build repository", "error", err)
	}
	if closeRepo != nil {
		defer closeRepo()
	}

	service := service.New(repo)

	r := chi.NewRouter()
	h := handler.New(config.BaseURL, service)

	r.Use(middlewares.Gzip, middlewares.Logging)

	r.Get("/ping", h.PingDatabase)
	r.Post("/api/shorten", h.CreateShortURLJson)
	r.Post("/", h.CreateShortURL)
	r.Get("/{id}", h.RedirectURL)

	logger.SugaredLogger.Infow(
		"Starting server",
		"addr", config.ServerAddress,
		"base_url", config.BaseURL,
		"database_dsn", config.DatabaseDSN,
		"file_storage_path", config.FileStoragePath,
	)

	err = http.ListenAndServe(config.ServerAddress, r)
	if err != nil && err != http.ErrServerClosed {
		logger.SugaredLogger.Fatalw("HTTP server failed: %v", err)
	}
}
