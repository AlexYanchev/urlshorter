package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newHandlerJSON(t *testing.T, originalJSON string) http.Handler {
	t.Helper()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}

		if string(body) != originalJSON {
			t.Errorf("expected decompressed body. want %s, got %s", originalJSON, body)
		}

		requestContentType := r.Header.Get("Content-Type")

		w.Header().Set("Content-Type", requestContentType)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(originalJSON))
	})
}

// клиент прислал сжатый gzip json - миддл распаковывает тело и получает обычный json

func TestGzip_ClientSendsGzipJson(t *testing.T) {
	originalJSON := `{"url":"https://example.com"}`

	handler := newHandlerJSON(t, originalJSON)

	var buf bytes.Buffer

	zw := gzip.NewWriter(&buf)
	zw.Write([]byte(originalJSON))
	zw.Close()

	req := httptest.NewRequest("POST", "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	Gzip(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

// клиент готов принимать сжатие - нужно сжать, добавить заголовки и отправить клиенту

func TestGzip_ClientAcceptsGzip(t *testing.T) {
	originalJSON := `{"url":"https://example.com"}`

	handler := newHandlerJSON(t, originalJSON)

	req := httptest.NewRequest("POST", "/", strings.NewReader(originalJSON))
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	Gzip(handler).ServeHTTP(rr, req)

	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", rr.Header().Get("Content-Type"))
	}

	var buf bytes.Buffer

	zw := gzip.NewWriter(&buf)
	zw.Write([]byte(originalJSON))
	zw.Close()
	
	zr, err := gzip.NewReader(rr.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer zr.Close()

	uncompressed, err := io.ReadAll(zr)
	if err != nil {
		t.Fatalf("failed to read uncompressed body: %v", err)
	}

	if string(uncompressed) != originalJSON {
		t.Fatalf("expected uncompressed body %s, got %s", originalJSON, uncompressed)
	}

	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected Content-Encoding gzip, got %s", rr.Header().Get("Content-Encoding"))
	}

	if rr.Header().Get("Content-Length") != "" {
		t.Errorf("expected empty Content-Length, got %s", rr.Header().Get("Content-Length"))
	}

	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", rr.Header().Get("Content-Type"))
	}

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

// клиент готов принимать сжатие, но тип контента не application/json text/html - не сжимаем

func TestGzip_ClientAcceptsGzipButContentTypeWrong(t *testing.T) {
	originalJSON := `{"url":"https://example.com"}`

	handler := newHandlerJSON(t, originalJSON)

	req := httptest.NewRequest("POST", "/", strings.NewReader(originalJSON))
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "text/plain")

	rr := httptest.NewRecorder()

	Gzip(handler).ServeHTTP(rr, req)

	body, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Errorf("failed to read body: %v", err)
	}

	trimmedBody := strings.TrimSuffix(string(body), "\n")

	if trimmedBody != originalJSON {
		t.Errorf("unexpected body %s, got %s", originalJSON, trimmedBody)
	}

	if rr.Header().Get("Content-Encoding") != "" {
		t.Errorf("expected empty Content-Encoding, got %s", rr.Header().Get("Content-Encoding"))
	}

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

// клиент не поддерживает сжатие - не сжимаем

func TestGzip_ClientNotSupportGzip(t *testing.T) {
	originalJSON := `{"url":"https://example.com"}`

	handler := newHandlerJSON(t, originalJSON)

	req := httptest.NewRequest("POST", "/", strings.NewReader(originalJSON))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	Gzip(handler).ServeHTTP(rr, req)

	body, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Errorf("failed to read body: %v", err)
	}

	if string(body) != originalJSON {
		t.Errorf("unexpected body %s, got %s", originalJSON, body)
	}

	if rr.Header().Get("Content-Encoding") != "" {
		t.Errorf("expected empty Content-Encoding, got %s", rr.Header().Get("Content-Encoding"))
	}

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

// сервер не может считать сжатые данные от клиента - нет паники, ошибка 4/5 

func TestGzip_InvalidGzipBody_ReturnsError(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        t.Error("handler was called, but should not be")
        w.WriteHeader(http.StatusOK)
    })

    invalidGzipData := []byte("this is not gzipped data, just plain text")
    req := httptest.NewRequest("POST", "/", bytes.NewReader(invalidGzipData))
    req.Header.Set("Content-Encoding", "gzip")
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()

    Gzip(handler).ServeHTTP(rr, req)

    if rr.Code == http.StatusOK {
        t.Errorf("expected error status (4xx/5xx), got %d", rr.Code)
    }
}