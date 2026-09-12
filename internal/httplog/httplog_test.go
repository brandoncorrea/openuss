package httplog_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"bwawan.com/openuss/internal/httplog"
	"bwawan.com/openuss/internal/logging"
	"bwawan.com/openuss/internal/logging/logtest"
)

func serve(t *testing.T, handler http.HandlerFunc, req *http.Request) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	logger, logs := logtest.New()
	response := httptest.NewRecorder()

	httplog.Middleware(logger)(handler).ServeHTTP(response, req)

	entry := logs.Find("inbound request")
	require.NotNil(t, entry)
	return response, entry
}

func TestLogsMethodPathStatusAndDuration(t *testing.T) {
	handler := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusCreated) }
	response, entry := serve(t, handler, httptest.NewRequest(http.MethodPut, "/flight_planning/v1/flight_plans/abc", nil))

	require.Equal(t, http.MethodPut, entry["method"])
	require.Equal(t, "/flight_planning/v1/flight_plans/abc", entry["path"])
	require.EqualValues(t, http.StatusCreated, entry["status"])
	require.Contains(t, entry, "duration")
	require.IsType(t, float64(0), entry["duration"])
	require.Equal(t, http.StatusCreated, response.Code)
}

func TestBareWriteCommitsOK(t *testing.T) {
	handler := func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("{}")) }
	response, entry := serve(t, handler, httptest.NewRequest(http.MethodGet, "/", nil))

	require.EqualValues(t, http.StatusOK, entry["status"])
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "{}", response.Body.String())
}

func TestHandlerThatWritesNothingIsOK(t *testing.T) {
	response, entry := serve(t, noopHandler, httptest.NewRequest(http.MethodGet, "/", nil))

	require.EqualValues(t, http.StatusOK, entry["status"])
	require.Equal(t, http.StatusOK, response.Code)
}

func TestSecondWriteHeaderDoesNotChangeLoggedStatus(t *testing.T) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.WriteHeader(http.StatusOK)
	}
	response, entry := serve(t, handler, httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.EqualValues(t, http.StatusInternalServerError, entry["status"])
}

func TestWriteHeaderAfterWriteDoesNotChangeLoggedStatus(t *testing.T) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("{}"))
		w.WriteHeader(http.StatusInternalServerError)
	}
	response, entry := serve(t, handler, httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusOK, response.Code)
	require.EqualValues(t, http.StatusOK, entry["status"])
}

func TestResponseControllerReachesUnderlyingWriter(t *testing.T) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		require.NoError(t, http.NewResponseController(w).Flush())
	}
	response, _ := serve(t, handler, httptest.NewRequest(http.MethodGet, "/", nil))

	require.True(t, response.Flushed)
}

func TestRecordCarriesRequestContextAttributes(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(logging.ContextWithAttrs(req.Context(), slog.String("request_id", "r-1")))

	_, entry := serve(t, noopHandler, req)

	require.Equal(t, "r-1", entry["request_id"])
}

func noopHandler(http.ResponseWriter, *http.Request) {}
