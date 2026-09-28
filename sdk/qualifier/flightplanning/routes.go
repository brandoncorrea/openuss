package flightplanning

import "net/http"

// TODO(gap): No inbound authentication or scope checks
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /flight_planning/v1/status", h.GetStatus)
	mux.HandleFunc("POST /flight_planning/v1/clear_area_requests", h.ClearAreaRequests)
	mux.HandleFunc("PUT /flight_planning/v1/flight_plans/{flight_plan_id}", h.PutFlightPlan)
	mux.HandleFunc("DELETE /flight_planning/v1/flight_plans/{flight_plan_id}", h.DeleteFlightPlan)
}
