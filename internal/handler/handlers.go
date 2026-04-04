package handler

import (
	"io"
	"net/http"
	"net/url"

	"github.com/AlexYanchev/urlshorter/internal/constants"
	"github.com/go-chi/chi/v5"
)

type URLService interface {
	CreateShortURL(originalURL string) (string, error)
	GetOriginalURL(id string) (string, error)
}

type Handler struct {
	service URLService
	baseAddressShortURL string
}

func New(baseAddressShortURL string, service URLService) *Handler {
	return &Handler{
		service: service,
		baseAddressShortURL: baseAddressShortURL,
	}
}

func (h *Handler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
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

	shortID, err := h.service.CreateShortURL(originalURL)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	host := h.baseAddressShortURL

	shortURL, err := url.JoinPath(host, shortID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (h *Handler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, constants.StatusIDNotProvided, http.StatusBadRequest)
		return
	}

	originalURL, err := h.service.GetOriginalURL(id)
	if err != nil {
		http.Error(w, constants.StatusURLNotFound, http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}