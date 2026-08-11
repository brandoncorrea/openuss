package scd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetVersion(t *testing.T) {
	response := httptest.NewRecorder()

	GetVersion(response, nil)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "application/json", response.Header().Get("Content-Type"))

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))

	systemVersion := map[string]string{
		"system_identity": "astm.f3548.v21",
		"system_version":  "",
	}
	require.Equal(t, systemVersion, body)
}
