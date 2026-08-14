package flightplanning

import (
	"net/http"

	"bwawan.com/openuss/internal/api"
)

func PutFlightPlan(w http.ResponseWriter, _ *http.Request) {
	api.WriteJSON(w, http.StatusOK, map[string]string{
		"planning_result":    "Rejected",
		"flight_plan_status": "NotPlanned",
	})
}
