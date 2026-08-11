package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bwawan.com/openuss/internal/logging/logtest"
)

func TestListenAddr(t *testing.T) {
	tests := []struct {
		name string
		port string
		want string
	}{
		{"defaults to port 80", "", ":80"},
		{"port 8080", "8080", ":8080"},
		{"port 3000", "3000", ":3000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PORT", tt.port)
			assert.Equal(t, tt.want, ResolveAddress())
		})
	}
}

func TestHandleShutdownReportsAFailedFlush(t *testing.T) {
	logger, rec := logtest.New()

	handleShutdown(func(context.Context) error {
		return errors.New("collector unreachable")
	}, logger)

	entry := rec.Find("tracing shutdown failed")
	require.NotNil(t, entry)
	require.Equal(t, "collector unreachable", entry["error"])
}

func TestHandleShutdownStaysQuietOnSuccess(t *testing.T) {
	logger, rec := logtest.New()
	handleShutdown(func(context.Context) error { return nil }, logger)
	require.Nil(t, rec.Find("tracing shutdown failed"))
}

func TestHandleShutdownGivesTheFlushALiveBudget(t *testing.T) {
	logger, _ := logtest.New()

	var (
		flushErr    error
		deadline    time.Time
		hasDeadline bool
	)

	handleShutdown(func(ctx context.Context) error {
		flushErr = ctx.Err()
		deadline, hasDeadline = ctx.Deadline()
		return nil
	}, logger)

	require.NoError(t, flushErr)
	require.True(t, hasDeadline)
	require.WithinDuration(t, time.Now().Add(flushTimeout), deadline, time.Second)
}
