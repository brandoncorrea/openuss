package flightplanning

import (
	"encoding/json"
	"net/http"
	"time"

	"bwawan.com/openuss/internal/api"
	"bwawan.com/openuss/internal/api/scdussv1"
)

type FlightPlanBasicInformation struct {
	UsageState  string              `json:"usage_state"`
	UasState    string              `json:"uas_state"`
	Area        []scdussv1.Volume4D `json:"area"`
	UtmId       *string             `json:"utm_id"`
	Description *string             `json:"description"`
}

type AstmF3548v21 struct {
	Priority int `json:"priority"`
}

type FlightPlan struct {
	BasicInformation FlightPlanBasicInformation `json:"basic_information"`
	Astm             AstmF3548v21               `json:"astm_f3548_21"`
}

func (flight *FlightPlan) StartTime() time.Time {
	// TODO(gap): Nothing validates a zero-area or multi-area flight plan
	return RFC3339(flight.BasicInformation.Area[0].TimeStart.Value)
}

func (flight *FlightPlan) EndTime() time.Time {
	// TODO(gap): Nothing validates a zero-area or multi-area flight plan
	return RFC3339(flight.BasicInformation.Area[0].TimeEnd.Value)
}

func RFC3339(s string) time.Time {
	// TODO(gap): Nothing validates a malformed timestamp
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

type PutFlightPlanBody struct {
	RequestId      scdussv1.UUIDv4Format `json:"request_id"`
	ExecutionStyle string                `json:"execution_style"`
	FlightPlan     FlightPlan            `json:"flight_plan"`
}

func PutFlightPlan(w http.ResponseWriter, r *http.Request) {
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
