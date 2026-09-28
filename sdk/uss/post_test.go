package uss_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNotifyOperationalIntent(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	handler := newHandler()
	handler.NotifyOperationalIntent(recorder, request)
	require.Equal(t, http.StatusNoContent, recorder.Code)
}
