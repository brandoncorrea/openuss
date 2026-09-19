package main_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	main "bwawan.com/openuss/cmd/openuss"
	"bwawan.com/openuss/internal/auth"
	"bwawan.com/openuss/internal/db"
	"bwawan.com/openuss/internal/dss"
	"bwawan.com/openuss/internal/flightplanning"
	"bwawan.com/openuss/internal/httpclient"
	"bwawan.com/openuss/internal/logging/logtest"
)

func TestListenAddr(t *testing.T) {
	tests := []struct {
		port string
		want string
	}{
		{"", ":8080"},
		{"8123", ":8123"},
		{"3000", ":3000"},
	}

	for _, tt := range tests {
		var name string
		if tt.port == "" {
			name = "Default"
		} else {
			name = "Port" + tt.port
		}
		t.Run(name, func(t *testing.T) {
			t.Setenv("PORT", tt.port)
			assert.Equal(t, tt.want, main.ResolveAddress())
		})
	}
}

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
	handler := main.NewHTTPHandler(db.NewInMemoryDB(), &flightplanning.Handler{}, logger)

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

func TestNewPlanningHandlerMissingUSSBaseURL(t *testing.T) {
	t.Setenv("DSS_BASE_URL", "http://dss.example.com")
	t.Setenv("USS_BASE_URL", "\r\n\t ")
	dummy, _ := auth.NewDummyOAuth("", "", nil)
	store := db.NewInMemoryDB()
	_, err := main.NewPlanningHandler(dummy, store)
	require.ErrorContains(t, err, "USS_BASE_URL is required")
}

func TestNewPlanningHandlerWithRealDSS(t *testing.T) {
	t.Setenv("DSS_BASE_URL", "http://dss.example.com:8080/blah")
	t.Setenv("USS_BASE_URL", "the-uss-base-url")
	dummy, _ := auth.NewDummyOAuth("", "", nil)
	store := db.NewInMemoryDB()
	handler, err := main.NewPlanningHandler(dummy, store)
	require.NoError(t, err)
	authority := handler.DSS.(*dss.DSS)
	require.Equal(t, "http://dss.example.com:8080/blah", authority.Host)
	require.Equal(t, dummy, authority.Client.TokenSource)
	require.Equal(t, httpclient.DefaultTimeout, authority.Client.HTTP.HTTP.Timeout)
	require.Equal(t, store, handler.DB)
	require.EqualValues(t, "the-uss-base-url", handler.USSBaseURL)
}

func TestNewPlanningHandlerWithMemoryDSS(t *testing.T) {
	t.Setenv("DSS_IMPL", "memory")
	t.Setenv("USS_BASE_URL", "the-uss-base-url")
	dummy, _ := auth.NewDummyOAuth("", "", nil)
	store := db.NewInMemoryDB()
	handler, err := main.NewPlanningHandler(dummy, store)
	require.NoError(t, err)
	require.IsType(t, &dss.InMemoryDSS{}, handler.DSS)
	require.Equal(t, store, handler.DB)
	require.EqualValues(t, "the-uss-base-url", handler.USSBaseURL)
}

func TestNewDummyTokenSource(t *testing.T) {
	t.Setenv("OAUTH_ENDPOINT", "http://oauth.local")
	t.Setenv("OAUTH_SUB", "the-oauth-subject")
	source, err := main.NewTokenSource()
	require.NoError(t, err)
	dummy := source.(*auth.DummyOAuth)
	require.Equal(t, "http://oauth.local", dummy.Endpoint.String())
	require.Equal(t, "the-oauth-subject", dummy.Subject)
}

func TestNewInMemoryTokenSource(t *testing.T) {
	t.Setenv("TOKEN_IMPL", "memory")
	source, err := main.NewTokenSource()
	require.NoError(t, err)
	require.IsType(t, auth.NewInMemoryTokenSource(), source)
}
