package dss

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/internal/auth"
	"github.com/stretchr/testify/assert"
)

func newDss(t *testing.T, handler http.HandlerFunc) *DSS {
	server := httptest.NewTestServer(t, http.HandlerFunc(handler))
	return &DSS{
		Host:        "http://dss.example.com",
		Audience:    "dss.example.com",
		TokenSource: auth.NewInMemoryTokenSource(),
		Client:      server.Client(),
	}
}

func assertNotCalledHandler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "expected handler not to be called")
	}
}
