package flightplanning

import (
	"context"
	"encoding/json/v2"
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
	if err != nil {
		return planningResponse(PlanningActivityResultRejected, intent)
	}

	h.Flights.Upsert(FlightPlanRecord{
		ID:       flightID,
		EntityID: intent.EntityID,
	})

	return planningResponse(PlanningActivityResultCompleted, intent)
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
	return scd.IntentParams{
		Volumes:  plan.BasicInformation.Area,
		State:    flightToIntentState(plan),
		Priority: scdussv1.Priority(plan.F3548.Priority),
	}
}

func planningResponse(result PlanningActivityResult, intent scd.OperationalIntent) FlightPlanResponse {
	return FlightPlanResponse{
		PlanningResult:   result,
		FlightPlanStatus: intentToFlightState(intent),
	}
}

func flightToIntentState(plan FlightPlan) scdussv1.OperationalIntentState {
	if plan.BasicInformation.UsageState == UsageStateInUse {
		return scdussv1.OperationalIntentState_Activated
	}
	return scdussv1.OperationalIntentState_Accepted
}

func intentToFlightState(intent scd.OperationalIntent) FlightPlanStatus {
	switch intent.State {
	case scdussv1.OperationalIntentState_Accepted:
		return FlightPlanStatusPlanned
	case scdussv1.OperationalIntentState_Activated:
		return FlightPlanStatusOkToFly
	default:
		return FlightPlanStatusNotPlanned
	}
}
