package flightplanning

import (
	"net/http"

	"bwawan.com/openuss/internal/api"
)

func (*Handler) DeleteFlightPlan(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	api.WriteJSON(w, http.StatusOK, map[string]any{
		"flight_plan_status": "Closed",
		"planning_result":    "Completed",
	})
}
