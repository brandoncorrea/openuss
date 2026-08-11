package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

// New returns a JSON logger writing to w, enriched with any attributes carried
// on the context by ContextWithAttrs.
func New(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(ContextHandler{
		Handler: slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level}),
	})
}

// Default returns the logger for a running OpenUSS process: JSON on stderr, at
// the level named by LOG_LEVEL.
func Default() *slog.Logger {
	return New(os.Stderr, resolveLevel(os.Getenv("LOG_LEVEL")))
}

func resolveLevel(name string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type attrsKey struct{}

// ContextWithAttrs returns a context carrying attrs, which ContextHandler adds
// to every record logged with it.
func ContextWithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	existing, _ := ctx.Value(attrsKey{}).([]slog.Attr)

	merged := make([]slog.Attr, 0, len(existing)+len(attrs))
	merged = append(merged, existing...)
	merged = append(merged, attrs...)

	return context.WithValue(ctx, attrsKey{}, merged)
}

// ContextHandler adds context-carried attributes to every record.
type ContextHandler struct {
	slog.Handler
}

func (h ContextHandler) Handle(ctx context.Context, rec slog.Record) error {
	if attrs, ok := ctx.Value(attrsKey{}).([]slog.Attr); ok {
		rec.AddAttrs(attrs...)
	}
	return h.Handler.Handle(ctx, rec)
}

func (h ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return ContextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h ContextHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return ContextHandler{Handler: h.Handler.WithGroup(name)}
}
