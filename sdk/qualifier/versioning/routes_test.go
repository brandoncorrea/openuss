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
			Path:    "/versioning/versions/FOO",
			Pattern: "GET /versioning/versions/{system_identity}",
			Handler: handler.GetVersion,
		},
	})
}
