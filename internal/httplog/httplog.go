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

func (rec *statusRecorder) WriteHeader(statusCode int) {
	if rec.statusCode == 0 {
		rec.statusCode = statusCode
	}
	rec.ResponseWriter.WriteHeader(statusCode)
}

func (rec *statusRecorder) Write(p []byte) (int, error) {
	if rec.statusCode == 0 {
		rec.statusCode = http.StatusOK
	}
	return rec.ResponseWriter.Write(p)
}

func (rec *statusRecorder) Unwrap() http.ResponseWriter {
	return rec.ResponseWriter
}

func (rec *statusRecorder) status() int {
	if rec.statusCode == 0 {
		return http.StatusOK
	}
	return rec.statusCode
}
