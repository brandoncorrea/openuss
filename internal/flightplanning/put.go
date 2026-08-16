package flightplanning

import (
	"encoding/json"
	"net/http"
	"time"

	"bwawan.com/openuss/internal/api"
	"bwawan.com/openuss/internal/api/scdussv1"
)

type PutFlightPlanBody struct {
	RequestId      scdussv1.UUIDv4Format `json:"request_id"`
	ExecutionStyle string                `json:"execution_style"`
	FlightPlan     FlightPlan            `json:"flight_plan"`
}

func (*Handler) PutFlightPlan(w http.ResponseWriter, r *http.Request) {
	var body PutFlightPlanBody

	// TODO(gap): What happens if malformed JSON is sent?
	json.NewDecoder(r.Body).Decode(&body)

	if isTooEager(body.FlightPlan) {
		api.WriteJSON(w, http.StatusOK, map[string]any{
			"activity_result":    "Rejected",
			"planning_result":    "Rejected",
			"flight_plan_status": "NotPlanned",
			// TODO(gap): Missing Fields: flight_id, includes_advisories, notes, queries(?), log_messages(?)
		})
	} else if hasEnded(body.FlightPlan) {
		api.WriteJSON(w, http.StatusOK, map[string]any{
			"activity_result":    "Rejected",
			"planning_result":    "Rejected",
			"flight_plan_status": "NotPlanned",
			// TODO(gap): Missing Fields: flight_id, includes_advisories, notes, queries(?), log_messages(?)
		})
	} else {
		api.WriteJSON(w, http.StatusOK, map[string]any{
			"planning_result":    "Completed",
			"flight_plan_status": "Planned",
			// TODO(gap): Missing Fields: activity_result, as_planned, flight_id, includes_advisories, queries(?), log_messages(?)
		})
	}
}

func isTooEager(flight FlightPlan) bool {
	return time.Now().Add(time.Hour * 24 * 30).Before(flight.StartTime())
}

func hasEnded(flight FlightPlan) bool {
	return flight.EndTime().Before(time.Now())
}
