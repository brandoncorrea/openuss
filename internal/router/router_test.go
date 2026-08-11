package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnknownPathIsNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/nope", nil)
	rec := httptest.NewRecorder()

	New().ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetVersion(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/versioning/versions/astm.f3548.v21", nil)
	rec := httptest.NewRecorder()

	New().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	systemVersion := map[string]string{
		"system_identity": "astm.f3548.v21",
		"system_version":  "",
	}
	require.Equal(t, systemVersion, body)
}
