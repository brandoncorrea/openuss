package utmclienttest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/sdk/auth"
	"bwawan.com/openuss/sdk/utmclient"
)

func NewClient(t *testing.T, handler http.HandlerFunc) *utmclient.Client {
	t.Helper()
	server := httptest.NewTestServer(t, handler)
	return utmclient.New(auth.NewInMemoryTokenSource(), server.Client())
}
