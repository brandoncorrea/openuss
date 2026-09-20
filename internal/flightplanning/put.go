package flightplanning

import (
	"context"
	"encoding/json/v2"
	"errors"
	"net/http"
	"uuid"

	"bwawan.com/openuss/internal/db"
	"bwawan.com/openuss/sdk/api"
	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/scd"
)

type PutFlightPlanBody struct {
	RequestID      scdussv1.UUIDv4Format `json:"request_id"`
	ExecutionStyle string                `json:"execution_style"`
	FlightPlan     FlightPlan            `json:"flight_plan"`
}

func (h *Handler) PutFlightPlan(w http.ResponseWriter, r *http.Request) {
	var body PutFlightPlanBody

	// TODO(gap): What happens if malformed JSON is sent?
	json.UnmarshalRead(r.Body, &body)

	id := uuid.MustParse(r.PathValue("flight_plan_id"))
	result := h.putOrRejectFlight(r.Context(), id, body.FlightPlan)
	api.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) putOrRejectFlight(
	ctx context.Context,
	flightID uuid.UUID,
	plan FlightPlan,
) map[string]string {
	existingFlight := h.Flights.Get(flightID)

	// TODO(gap): Validate flight_plan_id is a valid UUID, among other things
	intent, err := h.putIntent(ctx, existingFlight, toIntentParams(plan))
	if errors.Is(err, scd.ErrConflict) {
		return conflictResponse(existingFlight)
	}
	if errors.Is(err, scd.ErrRejected) {
		return rejectionResponse()
	}

	h.Flights.Upsert(db.FlightPlan{
		ID:       flightID,
		EntityID: intent.EntityID,
	})

	return successResponse(existingFlight)
}

func (h *Handler) putIntent(
	ctx context.Context,
	existingFlight *db.FlightPlan,
	params scd.IntentParams,
) (scd.OperationalIntent, error) {
	if existingFlight == nil {
		return h.SCD.CreateOperationalIntent(ctx, params)
	}
	return h.SCD.UpdateOperationalIntent(ctx, existingFlight.EntityID, params)
}

func toIntentParams(plan FlightPlan) scd.IntentParams {
	state := scdussv1.OperationalIntentState_Accepted
	if plan.BasicInformation.UsageState == "InUse" {
		state = scdussv1.OperationalIntentState_Activated
	}
	return scd.IntentParams{
		Volumes:  plan.BasicInformation.Area,
		State:    state,
		Priority: scdussv1.Priority(plan.F3548.Priority),
	}
}

func successResponse(existingFlight *db.FlightPlan) map[string]string {
	// TODO(gap): There's probably some input parameter this should be based off of
	status := "OkToFly"
	if existingFlight == nil {
		status = "Planned"
	}
	return map[string]string{
		"planning_result":    "Completed",
		"flight_plan_status": status,
		// TODO(gap): Missing Fields: activity_result, as_planned, flight_id, includes_advisories, queries(?), log_messages(?)
	}
}

func rejectionResponse() map[string]string {
	return map[string]string{
		"activity_result":    "Rejected",
		"planning_result":    "Rejected",
		"flight_plan_status": "NotPlanned",
		// TODO(gap): Missing Fields: flight_id, includes_advisories, notes, queries(?), log_messages(?)
	}
}

func conflictResponse(existingFlight *db.FlightPlan) map[string]string {
	// TODO(gap): This check is probably wrong
	status := "Planned"
	if existingFlight == nil {
		status = "NotPlanned"
	}
	return map[string]string{
		"activity_result":    "Rejected",
		"planning_result":    "Rejected",
		"flight_plan_status": status,
		// TODO(gap): Missing Fields: flight_id, includes_advisories, notes, queries(?), log_messages(?)
	}
}
