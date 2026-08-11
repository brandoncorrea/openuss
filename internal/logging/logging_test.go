package logging_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"bwawan.com/openuss/internal/logging"
	"bwawan.com/openuss/internal/logging/logtest"
)

func TestDefaultResolvesLevelFromEnv(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want slog.Level
	}{
		{"defaults to info", "", slog.LevelInfo},
		{"debug", "debug", slog.LevelDebug},
		{"warn", "warn", slog.LevelWarn},
		{"error", "error", slog.LevelError},
		{"case insensitive", "DEBUG", slog.LevelDebug},
		{"trims whitespace", "  error ", slog.LevelError},
		{"unrecognised falls back to info", "chatty", slog.LevelInfo},
	}

	ctx := context.Background()
	levels := []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LOG_LEVEL", tt.env)

			logger := logging.Default()
			for _, level := range levels {
				require.Equal(t, level >= tt.want, logger.Enabled(ctx, level),
					"level %s at LOG_LEVEL=%q", level, tt.env)
			}
		})
	}
}

func TestLogsAreJSONWithMessage(t *testing.T) {
	logger, rec := logtest.New()

	logger.InfoContext(context.Background(), "openuss listening", slog.String("addr", "[::]:80"))

	entry := rec.Find("openuss listening")
	require.NotNil(t, entry)
	require.Equal(t, "[::]:80", entry["addr"])
	require.Equal(t, "INFO", entry["level"])
}

func TestContextAttrsReachEveryRecord(t *testing.T) {
	logger, rec := logtest.New()

	ctx := logging.ContextWithAttrs(context.Background(), slog.String("flight_plan_id", "fp-1"))
	logger.InfoContext(ctx, "first")
	logger.InfoContext(ctx, "second")

	for _, msg := range []string{"first", "second"} {
		entry := rec.Find(msg)
		require.NotNil(t, entry)
		require.Equal(t, "fp-1", entry["flight_plan_id"])
	}
}

func TestContextAttrsAccumulate(t *testing.T) {
	logger, rec := logtest.New()

	ctx := logging.ContextWithAttrs(context.Background(), slog.String("flight_plan_id", "fp-1"))
	ctx = logging.ContextWithAttrs(ctx, slog.String("ovn", "ovn-9"))

	logger.InfoContext(ctx, "conflict evaluated")

	entry := rec.Find("conflict evaluated")
	require.NotNil(t, entry)
	require.Equal(t, "fp-1", entry["flight_plan_id"])
	require.Equal(t, "ovn-9", entry["ovn"])
}

func TestContextAttrsSurviveWith(t *testing.T) {
	logger, rec := logtest.New()

	ctx := logging.ContextWithAttrs(context.Background(), slog.String("trace_id", "abc"))
	logger.With(slog.String("surface", "flight_planning")).InfoContext(ctx, "derived")

	entry := rec.Find("derived")
	require.NotNil(t, entry)
	require.Equal(t, "abc", entry["trace_id"])
	require.Equal(t, "flight_planning", entry["surface"])
}

func TestContextAttrsSurviveWithGroup(t *testing.T) {
	logger, rec := logtest.New()

	ctx := logging.ContextWithAttrs(context.Background(), slog.String("trace_id", "abc"))
	logger.WithGroup("g").InfoContext(ctx, "grouped")

	entry := rec.Find("grouped")
	require.NotNil(t, entry)
	require.Equal(t, map[string]any{"trace_id": "abc"}, entry["g"])
}

func TestWithGroupIgnoresEmptyName(t *testing.T) {
	var buf bytes.Buffer
	handler := logging.ContextHandler{Handler: slog.NewJSONHandler(&buf, nil)}

	rec := slog.NewRecord(time.Now(), slog.LevelInfo, "m", 0)
	rec.AddAttrs(slog.String("k", "v"))

	require.NoError(t, handler.WithGroup("").Handle(context.Background(), rec))
	require.NotContains(t, buf.String(), `"":{`)
}

func TestContextAttrsDoNotLeakBetweenSiblings(t *testing.T) {
	logger, rec := logtest.New()

	// Three deep so the merged slice carries spare capacity.
	// a shallower chain reallocates on every branch and cannot catch append-in-place.
	base := logging.ContextWithAttrs(context.Background(), slog.String("a", "1"))
	base = logging.ContextWithAttrs(base, slog.String("b", "2"))
	base = logging.ContextWithAttrs(base, slog.String("c", "3"))

	left := logging.ContextWithAttrs(base, slog.String("only_left", "yes"))
	right := logging.ContextWithAttrs(base, slog.String("only_right", "yes"))

	logger.InfoContext(left, "left")
	logger.InfoContext(right, "right")

	entry := rec.Find("left")
	require.NotNil(t, entry)
	require.Equal(t, "yes", entry["only_left"])
	require.NotContains(t, entry, "only_right")
}
