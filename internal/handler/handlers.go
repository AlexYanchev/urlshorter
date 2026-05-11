package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/AlexYanchev/urlshorter/internal/constants"
	"github.com/AlexYanchev/urlshorter/internal/service"
	"github.com/go-chi/chi/v5"
)

type URLService interface {
	CreateShortURL(originalURL string) (string, error)
	CreateShortURLBatch(requests []service.BatchCreateRequest) ([]service.BatchCreateResult, error)
	GetOriginalURL(id string) (string, error)
	Ping() error
}

type Handler struct {
	service             URLService
	baseAddressShortURL string
}

type RequestJSON struct {
	URL string `json:"url"`
}

type ResponseJSON struct {
	Result string `json:"result"`
}

type BatchRequestJSON struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponseJSON struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func New(baseAddressShortURL string, service URLService) *Handler {
	return &Handler{
		service:             service,
		baseAddressShortURL: baseAddressShortURL,
	}
}

func (h *Handler) CreateShortURLJson(w http.ResponseWriter, r *http.Request) {
	var request RequestJSON

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, constants.StatusInvalidJSON, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if request.URL == "" {
		http.Error(w, constants.StatusInvalidURL, http.StatusBadRequest)
		return
	}

	if _, err := url.ParseRequestURI(request.URL); err != nil {
		http.Error(w, constants.StatusInvalidURL, http.StatusBadRequest)
		return
	}

	shortID, err := h.service.CreateShortURL(request.URL)
	if err != nil {
		var duplicateErr *service.DuplicateOriginalURLError
		
		if errors.As(err, &duplicateErr) {
			shortURL, joinErr := url.JoinPath(h.baseAddressShortURL, duplicateErr.ShortID)
			if joinErr != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			response := ResponseJSON{
				Result: shortURL,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			err = json.NewEncoder(w).Encode(response)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			
				return
			}

			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	host := h.baseAddressShortURL
	shortURL, err := url.JoinPath(host, shortID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := ResponseJSON{
		Result: shortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	
		return
	}
}

func (h *Handler) CreateShortURLBatch(w http.ResponseWriter, r *http.Request) {
	var request []BatchRequestJSON

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, constants.StatusInvalidJSON, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(request) == 0 {
		http.Error(w, constants.StatusInvalidJSON, http.StatusBadRequest)
		return
	}

	serviceRequest := make([]service.BatchCreateRequest, 0, len(request))
	for _, item := range request {
		if item.CorrelationID == "" || item.OriginalURL == "" {
			http.Error(w, constants.StatusInvalidURL, http.StatusBadRequest)
			return
		}

		if _, err := url.ParseRequestURI(item.OriginalURL); err != nil {
			http.Error(w, constants.StatusInvalidURL, http.StatusBadRequest)
			return
		}

		serviceRequest = append(serviceRequest, service.BatchCreateRequest{
			CorrelationID: item.CorrelationID,
			OriginalURL:   item.OriginalURL,
		})
	}

	serviceResponse, err := h.service.CreateShortURLBatch(serviceRequest)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := make([]BatchResponseJSON, 0, len(serviceResponse))
	for _, item := range serviceResponse {
		shortURL, err := url.JoinPath(h.baseAddressShortURL, item.ShortID)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response = append(response, BatchResponseJSON{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		
		return
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
		var duplicateErr *service.DuplicateOriginalURLError

		if errors.As(err, &duplicateErr) {
			shortURL, joinErr := url.JoinPath(h.baseAddressShortURL, duplicateErr.ShortID)
			if joinErr != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(shortURL))
			return
		}

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

func (h *Handler) PingDatabase(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Ping(); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
