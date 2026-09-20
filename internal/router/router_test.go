package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/internal/router"
	"github.com/stretchr/testify/require"
)

type textRoute struct {
	pattern string
	text    string
}

func (r textRoute) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc(r.pattern, func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(r.text))
	})
}

func serve(handler http.Handler, method, path string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(method, path, nil))
	return recorder
}

func TestServesTheRoutesOfEveryRegistrar(t *testing.T) {
	handler := router.New(
		textRoute{pattern: "GET /ping", text: "pong"},
		textRoute{pattern: "POST /echo", text: "echo"},
	)

	require.Equal(t, "pong", serve(handler, http.MethodGet, "/ping").Body.String())
	require.Equal(t, "echo", serve(handler, http.MethodPost, "/echo").Body.String())
}

func TestUnregisteredPathIsNotFound(t *testing.T) {
	handler := router.New(textRoute{pattern: "GET /ping", text: "pong"})

	require.Equal(t, http.StatusNotFound, serve(handler, http.MethodGet, "/blah").Code)
}
