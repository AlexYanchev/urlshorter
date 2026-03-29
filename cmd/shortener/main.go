package main

import (
	"log"
	"net/http"

	"github.com/AlexYanchev/urlshorter/internal/handler"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	h := handler.New()

	r.Post("/", h.CreateShortURL)
	r.Get("/{id}", h.RedirectURL)

	log.Println("Server starting on localhost:8080")
	
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}
