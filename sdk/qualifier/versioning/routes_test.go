package versioning_test

import (
	"net/http"
	"testing"

	"bwawan.com/openuss/sdk/internal/wiretest"
	"bwawan.com/openuss/sdk/qualifier/versioning"
)

func TestRoutes(t *testing.T) {
	handler := versioning.New()
	wiretest.RequireRouteRegistration(t, handler, []wiretest.RouteRegistration{
		{
			Method:  http.MethodGet,
			Path:    "/versioning/versions/astm.f3548.v21",
			Pattern: "GET /versioning/versions/astm.f3548.v21",
			Handler: handler.GetVersion,
		},
	})
}
