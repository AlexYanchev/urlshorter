package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"testing"

	"github.com/AlexYanchev/urlshorter/internal/constants"
	"github.com/go-chi/chi/v5"
)

func TestHandler_CreateShortURL(t *testing.T) {
	testBody := "http://example.ru"
	testPrefix := "http://"

	tests := []struct {
		name string
		method string
		body string
		expectedStatus int
		expectedBody string
	} {
		{
			name: "successful creation",
			method: http.MethodPost,
			body: testBody,
			expectedStatus: http.StatusCreated,
			expectedBody: testPrefix,
		},
		{
			name: "empty body",
			method: http.MethodPost,
			body: "",
			expectedStatus: http.StatusBadRequest,
			expectedBody: constants.StatusEmptyURL,
		},
		{
			name: "wrong method - GET",
			method: http.MethodGet,
			body: testBody,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody: constants.StatusMethodNotAllowed,
		},
		{
			name: "wrong method - PUT",
			method: http.MethodPut,
			body: testBody,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody: constants.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New()
			req := httptest.NewRequest(tt.method, "/", bytes.NewBufferString(tt.body))
			req.Host = "localhost:8080"

			rr := httptest.NewRecorder()

			h.CreateShortURL(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("wrong returned status code: got %v want %v", rr.Code, tt.expectedStatus)
			}

			body, err := io.ReadAll(rr.Body)
			if err != nil {
				t.Fatal(err)
			}

			if tt.expectedStatus == http.StatusCreated {
				contentType := rr.Header().Get("Content-Type")
				if contentType != "text/plain" {
					t.Errorf("wrong Content-Type: got %v, want text/plain", contentType)
				}

				if len(body) == 0 {
					t.Error("expected non empty body")
				}

				if !bytes.HasPrefix(body, []byte(testPrefix)) {
					t.Errorf("expected body prefix with %s, got %s", testPrefix, body)
				}
			} else {
				bodyString := strings.TrimSuffix(string(body), "\n")

				if bodyString != tt.expectedBody {
					t.Errorf("unexpected body: got %v want %v", bodyString, tt.expectedBody)
				}
			}
		})
	}
}

func TestHandler_RedirectURL(t *testing.T) {
	originalURL := "http://example.ru/test"

	h := New()
	r := chi.NewRouter()
	r.Post("/", h.CreateShortURL)
	r.Get("/{id}", h.RedirectURL)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(originalURL))
	req.Host = "localhost:8080"
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	body, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Error(err)
	}

	shortURL := string(body)
	parsedURL, err := url.Parse(shortURL)
	if err != nil {
		t.Error(err)
	}

	id := path.Base(parsedURL.Path)

	tests := []struct {
		name string
		method string
		id string
		expectedStatus int
		expectedLocation string
	} {
		{
			name: "success redirect",
			method: http.MethodGet,
			id: id,
			expectedStatus: http.StatusTemporaryRedirect,
			expectedLocation: originalURL,
		},
		{
			name: "ID non exist",
			method: http.MethodGet,
			id: "nonexist",
			expectedStatus: http.StatusNotFound,
			expectedLocation: "",
		},
		{
			name: "wrong method - POST",
			method: http.MethodPost,
			id: id,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedLocation: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/" + tt.id, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("returned wrong status code: got %v, want %v", rr.Code, tt.expectedStatus)
			}

			if rr.Code == http.StatusTemporaryRedirect {
				location := rr.Header().Get("Location")
				if location != tt.expectedLocation {
					t.Errorf("returned wrong Location: got %v, want %v", location, tt.expectedLocation)
				}
			}
		})
	}
}