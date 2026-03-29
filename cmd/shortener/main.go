package main

import (
	"log"
	"net/http"

	"github.com/AlexYanchev/urlshorter/internal/config"
	"github.com/AlexYanchev/urlshorter/internal/handler"
	"github.com/go-chi/chi/v5"
)

func main() {
	config := config.NewConfig()

	r := chi.NewRouter()
	h := handler.New(config.BaseShortURL)

	r.Post("/", h.CreateShortURL)
	r.Get("/{id}", h.RedirectURL)

	log.Printf("Server starting on %s. Base address for short url: %s", config.AppAddress, config.BaseShortURL)
	
	err := http.ListenAndServe(string(config.AppAddress), r)
	if err != nil {
		log.Fatal(err)
	}
}
