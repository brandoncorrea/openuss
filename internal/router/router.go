package router

import (
	"net/http"

	"bwawan.com/openuss/internal/scd/flightplanning"
	"bwawan.com/openuss/internal/scd/versioning"
)

// New returns a handler covering every route OpenUSS serves.
func New() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /versioning/versions/astm.f3548.v21", versioning.GetVersion)
	mux.HandleFunc("GET /flight_planning/v1/status", flightplanning.GetStatus)
	mux.HandleFunc("POST /flight_planning/v1/clear_area_requests", flightplanning.ClearArea)
	return mux
}
