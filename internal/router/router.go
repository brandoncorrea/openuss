package router

import "net/http"

type Versioning interface {
	GetVersion(http.ResponseWriter, *http.Request)
}

type FlightPlanning interface {
	GetStatus(http.ResponseWriter, *http.Request)
	ClearAreaRequests(http.ResponseWriter, *http.Request)
	PutFlightPlan(http.ResponseWriter, *http.Request)
	DeleteFlightPlan(http.ResponseWriter, *http.Request)
}

// New returns a handler covering every route OpenUSS serves.
func New(versioning Versioning, planning FlightPlanning) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /versioning/versions/astm.f3548.v21", versioning.GetVersion)
	mux.HandleFunc("GET /flight_planning/v1/status", planning.GetStatus)
	mux.HandleFunc("POST /flight_planning/v1/clear_area_requests", planning.ClearAreaRequests)
	mux.HandleFunc("PUT /flight_planning/v1/flight_plans/{flight_plan_id}", planning.PutFlightPlan)
	mux.HandleFunc("DELETE /flight_planning/v1/flight_plans/{flight_plan_id}", planning.DeleteFlightPlan)
	return mux
}
