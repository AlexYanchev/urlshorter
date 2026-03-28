package main

import (
	"log"
	"net/http"

	"github.com/AlexYanchev/urlshorter/internal/handler"
)

func main() {
	h := handler.New()

	http.HandleFunc("/", h.CreateShortURL)
	http.HandleFunc("/{id}", h.RedirectURL)

	log.Println("Server starting on localhost:8080")
	
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
