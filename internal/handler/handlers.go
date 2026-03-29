package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/AlexYanchev/urlshorter/internal/config"
	"github.com/AlexYanchev/urlshorter/internal/constants"
	"github.com/AlexYanchev/urlshorter/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *service.Service
	baseAddressShortURL config.AddressWithPortFlag
}

func New(baseAddressShortURL config.AddressWithPortFlag) *Handler {
	return &Handler{
		service: service.New(),
		baseAddressShortURL: baseAddressShortURL,
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
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	shortURL := fmt.Sprintf("%s://%s/%s", scheme, h.baseAddressShortURL, shortID)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (h *Handler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, constants.StatusMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, constants.StatusIDNotProvided, http.StatusBadRequest)
		return
	}

	originalURL, exists := h.service.GetOriginalURL(id)
	if !exists {
		http.Error(w, constants.StatusURLNotFound, http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}