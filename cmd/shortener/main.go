package main

import (
	"log"
	"net/http"

	"github.com/AlexYanchev/urlshorter/internal/config"
	"github.com/AlexYanchev/urlshorter/internal/handler"
	"github.com/AlexYanchev/urlshorter/internal/repository"
	"github.com/AlexYanchev/urlshorter/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	config := config.NewConfig()
	repo := repository.New()
	service := service.New(repo)

	r := chi.NewRouter()
	h := handler.New(config.BaseURL, service)

	r.Post("/", h.CreateShortURL)
	r.Get("/{id}", h.RedirectURL)

	log.Printf("Server starting on %s. Base address for short url: %s", config.ServerAddress, config.BaseURL)
	
	err := http.ListenAndServe(string(config.ServerAddress), r)
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server failed: %v", err)
	}

	log.Println("Server stopped")
}
