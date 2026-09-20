package flightplanning_test

import (
	"net/http"
	"testing"

	"bwawan.com/openuss/sdk/internal/wiretest"
	"bwawan.com/openuss/sdk/qualifier/flightplanning"
	"bwawan.com/openuss/sdk/scdtest"
)

func TestRoutes(t *testing.T) {
	handler := flightplanning.New(scdtest.Stub{}, flightplanning.NewInMemoryFlightStore())
	wiretest.RequireRouteRegistration(t, handler, []wiretest.RouteRegistration{
		{
			Method:  http.MethodGet,
			Path:    "/flight_planning/v1/status",
			Pattern: "GET /flight_planning/v1/status",
			Handler: handler.GetStatus,
		},
		{
			Method:  http.MethodPost,
			Path:    "/flight_planning/v1/clear_area_requests",
			Pattern: "POST /flight_planning/v1/clear_area_requests",
			Handler: handler.ClearAreaRequests,
		},
		{
			Method:  http.MethodPut,
			Path:    "/flight_planning/v1/flight_plans/FOO_ID",
			Pattern: "PUT /flight_planning/v1/flight_plans/{flight_plan_id}",
			Handler: handler.PutFlightPlan,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/flight_planning/v1/flight_plans/FOO_ID",
			Pattern: "DELETE /flight_planning/v1/flight_plans/{flight_plan_id}",
			Handler: handler.DeleteFlightPlan,
		},
	})
}
