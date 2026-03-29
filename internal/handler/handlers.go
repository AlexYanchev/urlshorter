package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/AlexYanchev/urlshorter/internal/constants"
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
		http.Error(w, constants.StatusMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, constants.StatusFailedReadBody, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalURL := string(body)
	if originalURL == "" {
		http.Error(w, constants.StatusEmptyURL, http.StatusBadRequest)
		return
	}

	shortID := h.service.CreateShortURL(originalURL)
	shortURL := fmt.Sprintf("http://%s/%s", r.Host, shortID)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (h *Handler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, constants.StatusMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		http.Error(w, constants.StatusIDNotProvided, http.StatusBadRequest)
		return
	}

	originalURL, exists := h.service.GetOriginalURL(path)
	if !exists {
		http.Error(w, constants.StatusURLNotFound, http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}