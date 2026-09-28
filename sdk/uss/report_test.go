package uss_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMakeUSSReport(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	handler := newHandler()
	handler.MakeUSSReport(recorder, request)
	require.Equal(t, http.StatusCreated, recorder.Code)
}
