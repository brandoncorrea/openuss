package httplog

import (
	"log/slog"
	"net/http"
	"time"
)

func Middleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w}

			next.ServeHTTP(rec, r)

			logger.InfoContext(r.Context(), "inbound request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status()),
				slog.Duration("duration", time.Since(start)))
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (s *statusRecorder) WriteHeader(statusCode int) {
	if s.statusCode == 0 {
		s.statusCode = statusCode
	}
	s.ResponseWriter.WriteHeader(statusCode)
}

func (s *statusRecorder) Write(p []byte) (int, error) {
	if s.statusCode == 0 {
		s.statusCode = http.StatusOK
	}
	return s.ResponseWriter.Write(p)
}

func (s *statusRecorder) Unwrap() http.ResponseWriter {
	return s.ResponseWriter
}

func (s *statusRecorder) status() int {
	if s.statusCode == 0 {
		return http.StatusOK
	}
	return s.statusCode
}
