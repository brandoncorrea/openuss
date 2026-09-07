package flightplanning

import (
	"encoding/json/v2"
	"net/http"
	"slices"
	"time"
	"uuid"

	"bwawan.com/openuss/internal/api"
	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/db"
)

type PutFlightPlanBody struct {
	RequestId      scdussv1.UUIDv4Format `json:"request_id"`
	ExecutionStyle string                `json:"execution_style"`
	FlightPlan     FlightPlan            `json:"flight_plan"`
}

func (handler *Handler) PutFlightPlan(w http.ResponseWriter, r *http.Request) {
	var body PutFlightPlanBody

	// TODO(gap): What happens if malformed JSON is sent?
	json.UnmarshalRead(r.Body, &body)

	// TODO(gap): Validate flight_plan_id is a valid UUID
	if isTooEager(body.FlightPlan) || hasEnded(body.FlightPlan) || hasAnyIntent(handler) {
		writeRejection(w)
	} else {
		intent := scdussv1.PutOperationalIntentReferenceParameters{
			Extents:    body.FlightPlan.BasicInformation.Area,
			State:      scdussv1.OperationalIntentState_Accepted,
			UssBaseUrl: "http://host.docker.internal:8080",
		}

		// TODO(gap): What happens if the DSS call results in an error?
		// TODO(next): DSS returns a conflict - should address the TODO(gap) above
		result, _ := handler.DSS.CreateOperationalIntentReference(r.Context(), scdussv1.EntityID(uuid.New().String()), intent)

		timeStart, _ := time.Parse(time.RFC3339Nano, result.OperationalIntentReference.TimeStart.Value)
		timeEnd, _ := time.Parse(time.RFC3339Nano, result.OperationalIntentReference.TimeEnd.Value)
		handler.DB.SaveIntent(db.OperationalIntent{
			EntityID:        result.OperationalIntentReference.Id,
			Manager:         result.OperationalIntentReference.Manager,
			UssAvailability: result.OperationalIntentReference.UssAvailability,
			Version:         result.OperationalIntentReference.Version,
			State:           result.OperationalIntentReference.State,
			Ovn:             *result.OperationalIntentReference.Ovn,
			TimeStart:       timeStart,
			TimeEnd:         timeEnd,
			UssBaseUrl:      result.OperationalIntentReference.UssBaseUrl,
			SubscriptionId:  result.OperationalIntentReference.SubscriptionId,
			Volumes:         intent.Extents,
		})
		handler.DB.SaveFlight(db.FlightPlan{
			Id:       uuid.MustParse(r.PathValue("flight_plan_id")),
			EntityID: result.OperationalIntentReference.Id,
		})
		api.WriteJSON(w, http.StatusOK, map[string]any{
			"planning_result":    "Completed",
			"flight_plan_status": "Planned",
			// TODO(gap): Missing Fields: activity_result, as_planned, flight_id, includes_advisories, queries(?), log_messages(?)
		})
	}
}

func writeRejection(w http.ResponseWriter) {
	api.WriteJSON(w, http.StatusOK, map[string]any{
		"activity_result":    "Rejected",
		"planning_result":    "Rejected",
		"flight_plan_status": "NotPlanned",
		// TODO(gap): Missing Fields: flight_id, includes_advisories, notes, queries(?), log_messages(?)
	})
}

func hasAnyIntent(handler *Handler) bool {
	return len(slices.Collect(handler.DB.GetAllIntents())) > 0
}

func isTooEager(flight FlightPlan) bool {
	return time.Now().Add(time.Hour * 24 * 30).Before(flight.StartTime())
}

func hasEnded(flight FlightPlan) bool {
	return flight.EndTime().Before(time.Now())
}
