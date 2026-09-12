package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bwawan.com/openuss/internal/auth"
	"bwawan.com/openuss/internal/db"
	"bwawan.com/openuss/internal/dss"
	"bwawan.com/openuss/internal/flightplanning"
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

func serveThroughNewHandler(t *testing.T, method, target string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	logger, logs := logtest.New()
	handler := newHandler(db.NewInMemoryDB(), &flightplanning.Handler{}, logger)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(method, target, nil))
	return response, logs.Find("inbound request")
}

func TestNewHandlerLogsRoutedRequests(t *testing.T) {
	response, entry := serveThroughNewHandler(t, http.MethodGet, "/versioning/versions/astm.f3548.v21")

	require.Equal(t, http.StatusOK, response.Code)
	require.NotNil(t, entry, "expected the request to be logged")
	require.Equal(t, "/versioning/versions/astm.f3548.v21", entry["path"])
	require.EqualValues(t, http.StatusOK, entry["status"])
}

// The middleware wraps the mux, not the individual routes, so unknown paths
// are logged too.
func TestNewHandlerLogsUnroutedRequests(t *testing.T) {
	response, entry := serveThroughNewHandler(t, http.MethodGet, "/nope")

	require.Equal(t, http.StatusNotFound, response.Code)
	require.NotNil(t, entry, "expected the 404 to be logged")
	require.EqualValues(t, http.StatusNotFound, entry["status"])
}

func TestNewPlanningHandlerMissingUssBaseUrl(t *testing.T) {
	t.Setenv("DSS_BASE_URL", "http://dss.example.com")
	t.Setenv("USS_BASE_URL", "\r\n\t ")
	dummy, _ := auth.NewDummyOAuth("", "", nil)
	db := db.NewInMemoryDB()
	_, err := newPlanningHandler(dummy, db)
	require.ErrorContains(t, err, "USS_BASE_URL is required")
}

func TestNewPlanningHandlerWithRealDSS(t *testing.T) {
	t.Setenv("DSS_BASE_URL", "http://dss.example.com:8080/blah")
	t.Setenv("USS_BASE_URL", "the-uss-base-url")
	dummy, _ := auth.NewDummyOAuth("", "", nil)
	db := db.NewInMemoryDB()
	handler, err := newPlanningHandler(dummy, db)
	require.NoError(t, err)
	dss := handler.DSS.(*dss.DSS)
	require.Equal(t, http.DefaultClient, dss.Client)
	require.Equal(t, "http://dss.example.com:8080/blah", dss.Host)
	require.Equal(t, "dss.example.com", dss.Audience)
	require.Equal(t, dummy, dss.TokenSource)
	require.Equal(t, db, handler.DB)
	require.EqualValues(t, "the-uss-base-url", handler.UssBaseUrl)
}

func TestNewPlanningHandlerWithMemoryDSS(t *testing.T) {
	t.Setenv("DSS_IMPL", "memory")
	t.Setenv("USS_BASE_URL", "the-uss-base-url")
	dummy, _ := auth.NewDummyOAuth("", "", nil)
	db := db.NewInMemoryDB()
	handler, err := newPlanningHandler(dummy, db)
	require.NoError(t, err)
	require.IsType(t, &dss.InMemoryDSS{}, handler.DSS)
	require.Equal(t, db, handler.DB)
	require.EqualValues(t, "the-uss-base-url", handler.UssBaseUrl)
}

func TestNewDummyTokenSource(t *testing.T) {
	t.Setenv("OAUTH_ENDPOINT", "http://oauth.local")
	t.Setenv("OAUTH_SUB", "the-oauth-subject")
	source, err := newTokenSource()
	require.NoError(t, err)
	dummy := source.(*auth.DummyOAuth)
	require.Equal(t, "http://oauth.local", dummy.Endpoint.String())
	require.Equal(t, "the-oauth-subject", dummy.Subject)
}

func TestNewInMemoryTokenSource(t *testing.T) {
	t.Setenv("TOKEN_IMPL", "memory")
	source, err := newTokenSource()
	require.NoError(t, err)
	require.IsType(t, auth.NewInMemoryTokenSource(), source)
}
