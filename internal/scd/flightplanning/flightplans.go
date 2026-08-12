package flightplanning

import (
	"net/http"

	"bwawan.com/openuss/internal/util"
)

func PutFlightPlan(w http.ResponseWriter, _ *http.Request) {
	util.WriteJSON(w, map[string]any{
		"planning_result":    "Rejected",
		"flight_plan_status": "NotPlanned",
	})
}
