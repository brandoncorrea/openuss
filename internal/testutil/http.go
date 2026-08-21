package testutil

import (
	"encoding/json/v2"
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
	require.NoError(t, json.UnmarshalRead(recorder.Body, &body))
	require.Equal(t, expected, body)
}

type BadTransport struct {
	Error error
}

func (transport *BadTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, transport.Error
}

func NewErrorClient(err error) *http.Client {
	return &http.Client{Transport: &BadTransport{Error: err}}
}
