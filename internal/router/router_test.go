package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type FakeVersioning struct{}
type FakeFlightPlanning struct{}

func RespondText(w http.ResponseWriter, content string) {
	w.Write([]byte(content))
}

func (*FakeVersioning) GetVersion(w http.ResponseWriter, r *http.Request) {
	RespondText(w, "GetVersion")
}

func (*FakeFlightPlanning) GetStatus(w http.ResponseWriter, r *http.Request) {
	RespondText(w, "GetStatus")
}

func (*FakeFlightPlanning) ClearAreaRequests(w http.ResponseWriter, r *http.Request) {
	RespondText(w, "ClearAreaRequests")
}

func (*FakeFlightPlanning) PutFlightPlan(w http.ResponseWriter, r *http.Request) {
	RespondText(w, "PutFlightPlan: "+r.PathValue("flight_plan_id"))
}

func (*FakeFlightPlanning) DeleteFlightPlan(w http.ResponseWriter, r *http.Request) {
	RespondText(w, "DeleteFlightPlan: "+r.PathValue("flight_plan_id"))
}

func NewFakeHandler() http.Handler {
	return New(&FakeVersioning{}, &FakeFlightPlanning{})
}

func TestRoutes(t *testing.T) {
	type Route struct {
		Method string
		Path   string
		Result string
	}

	handler := NewFakeHandler()

	for _, route := range []Route{
		{
			Method: http.MethodGet,
			Path:   "/versioning/versions/astm.f3548.v21",
			Result: "GetVersion",
		},
		{
			Method: http.MethodGet,
			Path:   "/flight_planning/v1/status",
			Result: "GetStatus",
		},
		{
			Method: http.MethodPost,
			Path:   "/flight_planning/v1/clear_area_requests",
			Result: "ClearAreaRequests",
		},
		{
			Method: http.MethodPut,
			Path:   "/flight_planning/v1/flight_plans/FOO_ID",
			Result: "PutFlightPlan: FOO_ID",
		},
		{
			Method: http.MethodDelete,
			Path:   "/flight_planning/v1/flight_plans/BAR_ID",
			Result: "DeleteFlightPlan: BAR_ID",
		},
	} {
		t.Run(route.Method+" "+route.Path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(route.Method, route.Path, nil)
			handler.ServeHTTP(recorder, request)
			require.Equal(t, route.Result, recorder.Body.String())
		})
	}
}

func TestNotFoundHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/blah", nil)
	rec := httptest.NewRecorder()
	NewFakeHandler().ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}
