package main

import (
	"log"
	"net/http"

	"github.com/AlexYanchev/urlshorter/internal/config"
	"github.com/AlexYanchev/urlshorter/internal/handler"
	"github.com/AlexYanchev/urlshorter/internal/logger"
	"github.com/AlexYanchev/urlshorter/internal/middlewares"
	"github.com/AlexYanchev/urlshorter/internal/repository"
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
	repo := repository.New()
	service := service.New(repo)

	r := chi.NewRouter()
	h := handler.New(config.BaseURL, service)

	r.Use(middlewares.Gzip, middlewares.Logging)

	r.Post("/api/shorten", h.CreateShortURLJson)
	r.Post("/", h.CreateShortURL)
	r.Get("/{id}", h.RedirectURL)

	logger.SugaredLogger.Infow(
        "Starting server",
        "addr", config.ServerAddress,
        "base_url", config.BaseURL,
    )

	err = http.ListenAndServe(config.ServerAddress, r)
	if err != nil && err != http.ErrServerClosed {
		logger.SugaredLogger.Fatalw("HTTP server failed: %v", err)
	}
}