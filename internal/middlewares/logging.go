package middlewares

import (
	"net/http"
	"time"

	"github.com/AlexYanchev/urlshorter/internal/logger"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size	   int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseWriter) Write(b []byte) (int, error) {
    size, err := r.ResponseWriter.Write(b) 
    r.size += size

    return size, err
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := &responseWriter{w, http.StatusOK, 0}
		next.ServeHTTP(ww, r)

		duration := time.Since(start)

		logger.SugaredLogger.Infow(
			"REQUEST", 
			"uri", r.RequestURI, 
			"method", r.Method, 
			"duration_nano", duration.Nanoseconds(),
		)
		logger.SugaredLogger.Infow(
			"RESPONSE", 
			"status", ww.statusCode, 
			"size", ww.size,
		)
	})
}