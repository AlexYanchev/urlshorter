package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/AlexYanchev/urlshorter/internal/service"
)

type Handler struct {
	service *service.Service
}

func New() *Handler {
	return &Handler{
		service: service.New(),
	}
}

func (h *Handler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalURL := string(body)
	if originalURL == "" {
		http.Error(w, "URL cannot be empty", http.StatusBadRequest)
	}

	shortID := h.service.CreateShortURL(originalURL)
	shortURL := fmt.Sprintf("http://%s/%s", r.Host, shortID)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (h *Handler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		http.Error(w, "ID not provided", http.StatusBadRequest)
		return
	}

	originalURL, exists := h.service.GetOriginalURL(path)
	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}