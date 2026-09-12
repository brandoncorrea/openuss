package dss

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/internal/auth"
	"bwawan.com/openuss/internal/utmclient"
)

func newDss(t *testing.T, handler http.HandlerFunc) *DSS {
	server := httptest.NewTestServer(t, http.HandlerFunc(handler))
	return &DSS{
		Host:   "http://dss.example.com",
		Client: utmclient.New(auth.NewInMemoryTokenSource(), server.Client()),
	}
}
