package flightplanning

import (
	"context"
	"encoding/json/v2"
	"errors"
	"net/http"
	"uuid"

	"bwawan.com/openuss/sdk/api"
	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/scd"
)

func (h *Handler) PutFlightPlan(w http.ResponseWriter, r *http.Request) {
	var body UpsertFlightPlanRequest

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
) FlightPlanResponse {
	existingFlight := h.Flights.Get(flightID)

	// TODO(gap): Validate flight_plan_id is a valid UUID, among other things
	intent, err := h.putIntent(ctx, existingFlight, toIntentParams(plan))
	if errors.Is(err, scd.ErrConflict) {
		return conflictResponse(existingFlight)
	}
	if errors.Is(err, scd.ErrRejected) {
		return conflictResponse(nil)
	}

	h.Flights.Upsert(FlightPlanRecord{
		ID:       flightID,
		EntityID: intent.EntityID,
	})

	return successResponse(existingFlight)
}

func (h *Handler) putIntent(
	ctx context.Context,
	existingFlight *FlightPlanRecord,
	params scd.IntentParams,
) (scd.OperationalIntent, error) {
	if existingFlight == nil {
		return h.SCD.CreateOperationalIntent(ctx, params)
	}
	return h.SCD.UpdateOperationalIntent(ctx, existingFlight.EntityID, params)
}

func toIntentParams(plan FlightPlan) scd.IntentParams {
	state := scdussv1.OperationalIntentState_Accepted
	if plan.BasicInformation.UsageState == UsageStateInUse {
		state = scdussv1.OperationalIntentState_Activated
	}
	return scd.IntentParams{
		Volumes:  plan.BasicInformation.Area,
		State:    state,
		Priority: scdussv1.Priority(plan.F3548.Priority),
	}
}

func successResponse(existingFlight *FlightPlanRecord) FlightPlanResponse {
	// TODO(gap): There's probably some input parameter this should be based off of
	status := FlightPlanStatusOkToFly
	if existingFlight == nil {
		status = FlightPlanStatusPlanned
	}
	return FlightPlanResponse{
		PlanningResult:   PlanningActivityResultCompleted,
		FlightPlanStatus: status,
	}
}

func conflictResponse(existingFlight *FlightPlanRecord) FlightPlanResponse {
	// TODO(gap): This check is probably wrong
	status := FlightPlanStatusPlanned
	if existingFlight == nil {
		status = FlightPlanStatusNotPlanned
	}
	return FlightPlanResponse{
		PlanningResult:   PlanningActivityResultRejected,
		FlightPlanStatus: status,
	}
}
