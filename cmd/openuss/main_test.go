package main_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	main "bwawan.com/openuss/cmd/openuss"
	"bwawan.com/openuss/internal/logging/logtest"
	"bwawan.com/openuss/sdk/qualifier/flightplanning"
	"bwawan.com/openuss/sdk/scd"
)

func TestHandleShutdownReportsAFailedFlush(t *testing.T) {
	logger, rec := logtest.New()

	main.HandleShutdown(func(context.Context) error {
		return errors.New("collector unreachable")
	}, logger)

	entry := rec.Find("tracing shutdown failed")
	require.NotNil(t, entry)
	require.Equal(t, "collector unreachable", entry["error"])
}

func TestHandleShutdownStaysQuietOnSuccess(t *testing.T) {
	logger, rec := logtest.New()
	main.HandleShutdown(func(context.Context) error { return nil }, logger)
	require.Nil(t, rec.Find("tracing shutdown failed"))
}

func TestHandleShutdownGivesTheFlushALiveBudget(t *testing.T) {
	logger, _ := logtest.New()

	var (
		flushErr    error
		deadline    time.Time
		hasDeadline bool
	)

	main.HandleShutdown(func(ctx context.Context) error {
		flushErr = ctx.Err()
		deadline, hasDeadline = ctx.Deadline()
		return nil
	}, logger)

	require.NoError(t, flushErr)
	require.True(t, hasDeadline)
	require.WithinDuration(t, time.Now().Add(main.FlushTimeout), deadline, time.Second)
}

func serveThroughNewHTTPHandler(t *testing.T, method, target string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	logger, logs := logtest.New()
	handler := main.NewHTTPHandler(main.Services{
		Intents:  scd.NewInMemoryIntentStore(),
		Planning: &flightplanning.Handler{},
	}, logger)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(method, target, nil))
	return response, logs.Find("inbound request")
}

func TestNewHandlerLogsRoutedRequests(t *testing.T) {
	response, entry := serveThroughNewHTTPHandler(t, http.MethodGet, "/versioning/versions/astm.f3548.v21")

	require.Equal(t, http.StatusOK, response.Code)
	require.NotNil(t, entry, "expected the request to be logged")
	require.Equal(t, "/versioning/versions/astm.f3548.v21", entry["path"])
	require.EqualValues(t, http.StatusOK, entry["status"])
}

// The middleware wraps the mux, not the individual routes, so unknown paths
// are logged too.
func TestNewHandlerLogsUnroutedRequests(t *testing.T) {
	response, entry := serveThroughNewHTTPHandler(t, http.MethodGet, "/nope")

	require.Equal(t, http.StatusNotFound, response.Code)
	require.NotNil(t, entry, "expected the 404 to be logged")
	require.EqualValues(t, http.StatusNotFound, entry["status"])
}
