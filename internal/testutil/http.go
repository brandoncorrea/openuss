package testutil

import (
	"encoding/json"
	"mime"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func RequireJSON(t *testing.T, recorder *httptest.ResponseRecorder, expected map[string]any) {
	t.Helper()
	require.Equal(t, http.StatusOK, recorder.Code)

	mediaType, _, err := mime.ParseMediaType(recorder.Header().Get("Content-Type"))
	require.NoError(t, err)
	require.Equal(t, "application/json", mediaType)

	var body map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, expected, body)
}
