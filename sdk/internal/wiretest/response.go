package wiretest

import (
	"encoding/json/v2"
	"mime"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func RequireJSON[T any](t *testing.T, recorder *httptest.ResponseRecorder, expected T) {
	t.Helper()
	require.Equal(t, http.StatusOK, recorder.Code)

	mediaType, _, err := mime.ParseMediaType(recorder.Header().Get("Content-Type"))
	require.NoError(t, err)
	require.Equal(t, "application/json", mediaType)

	var body T
	require.NoError(t, json.UnmarshalRead(recorder.Body, &body))
	require.Equal(t, expected, body)
}
