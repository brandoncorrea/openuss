package flightplanning

import (
	"context"
	"encoding/json/v2"
	"errors"
	"net/http"
	"slices"
	"time"
	"uuid"

	"bwawan.com/openuss/internal/api"
	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/db"
	"bwawan.com/openuss/internal/dss"
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
	status, result := h.putOrRejectFlight(r.Context(), id, body.FlightPlan)
	api.WriteJSON(w, status, result)
}

func (h *Handler) putOrRejectFlight(
	ctx context.Context,
	flightID uuid.UUID,
	plan FlightPlan,
) (int, any) {
	if isInvalidFlight(plan) {
		return rejectionResponse()
	}

	existingFlight := h.DB.GetFlight(flightID)
	entityID, ovn := h.findIDsForExistingFlightPlan(existingFlight)

	// TODO(gap): This assumes everything overlaps
	if h.hasAnyOtherIntent(entityID) {
		return rejectionResponse()
	}

	putParams := h.createPutRequestParams(plan)

	// TODO(gap): What happens if the DSS call results in a non-conflict error?
	result, err := h.DSS.PutOperationalIntentReference(ctx, entityID, ovn, putParams)
	if conflict, ok := errors.AsType[dss.AirspaceConflictError](err); ok {
		missing := (*conflict.MissingOperationalIntents)[0]
		details, _ := h.Peer.GetOperationalIntentDetails(ctx, missing.UssBaseUrl, missing.Id)
		if isLowerPriority(details) {
			// TODO(gap): This check is probably wrong
			status := "NotPlanned"
			if ovn != nil {
				status = "Planned"
			}
			return http.StatusOK, map[string]any{
				"activity_result":    "Rejected",
				"planning_result":    "Rejected",
				"flight_plan_status": status,
				// TODO(gap): Missing Fields: flight_id, includes_advisories, notes, queries(?), log_messages(?)
			}
		}
		putParams.Key = &scdussv1.Key{*details.OperationalIntent.Reference.Ovn}
		result, err = h.DSS.PutOperationalIntentReference(ctx, entityID, ovn, putParams)
	}

	intent := h.saveOperationalIntent(plan, result)
	h.saveFlightPlan(flightID, intent)

	return http.StatusOK, map[string]any{
		"planning_result":    "Completed",
		"flight_plan_status": flightPlanStatus(ovn),
		// TODO(gap): Missing Fields: activity_result, as_planned, flight_id, includes_advisories, queries(?), log_messages(?)
	}
}

func isLowerPriority(details scdussv1.GetOperationalIntentDetailsResponse) bool {
	return details.OperationalIntent.Details.Priority != nil &&
		*details.OperationalIntent.Details.Priority == 100
}

func isInvalidFlight(plan FlightPlan) bool {
	// TODO(gap): Validate flight_plan_id is a valid UUID, among other things
	return isTooEager(plan) || hasEnded(plan)
}

func (h *Handler) findIDsForExistingFlightPlan(flight *db.FlightPlan) (scdussv1.EntityID, *scdussv1.EntityOVN) {
	if flight == nil {
		return scdussv1.EntityID(uuid.New().String()), nil
	}
	intent := h.DB.GetIntent(flight.EntityID)
	return intent.EntityID, new(intent.OVN)
}

func (h *Handler) createPutRequestParams(plan FlightPlan) scdussv1.PutOperationalIntentReferenceParameters {
	return scdussv1.PutOperationalIntentReferenceParameters{
		Extents:    plan.BasicInformation.Area,
		State:      flightPlanState(plan),
		UssBaseUrl: h.USSBaseURL,
		NewSubscription: &scdussv1.ImplicitSubscriptionParameters{
			// TODO(gap): This probably needs to be a proper URL
			UssBaseUrl: scdussv1.SubscriptionUssBaseURL("x"),
		},
	}
}

func flightPlanState(plan FlightPlan) scdussv1.OperationalIntentState {
	if plan.BasicInformation.UsageState == "InUse" {
		return scdussv1.OperationalIntentState_Activated
	}
	return scdussv1.OperationalIntentState_Accepted
}

func (h *Handler) saveOperationalIntent(
	plan FlightPlan,
	result scdussv1.ChangeOperationalIntentReferenceResponse,
) db.OperationalIntent {
	timeStart, _ := time.Parse(time.RFC3339Nano, result.OperationalIntentReference.TimeStart.Value)
	timeEnd, _ := time.Parse(time.RFC3339Nano, result.OperationalIntentReference.TimeEnd.Value)
	intent := db.OperationalIntent{
		EntityID:        result.OperationalIntentReference.Id,
		Manager:         result.OperationalIntentReference.Manager,
		USSAvailability: result.OperationalIntentReference.UssAvailability,
		Version:         result.OperationalIntentReference.Version,
		Priority:        scdussv1.Priority(plan.F3548.Priority),
		State:           result.OperationalIntentReference.State,
		OVN:             *result.OperationalIntentReference.Ovn,
		TimeStart:       timeStart,
		TimeEnd:         timeEnd,
		USSBaseURL:      result.OperationalIntentReference.UssBaseUrl,
		SubscriptionID:  result.OperationalIntentReference.SubscriptionId,
		Volumes:         plan.BasicInformation.Area,
	}
	h.DB.SaveIntent(intent)
	return intent
}

func (h *Handler) saveFlightPlan(id uuid.UUID, intent db.OperationalIntent) {
	h.DB.SaveFlight(db.FlightPlan{
		ID:       id,
		EntityID: intent.EntityID,
	})
}

func flightPlanStatus(ovn *scdussv1.EntityOVN) string {
	// TODO(gap): There's probably some input parameter this should be based off of
	if ovn == nil {
		return "Planned"
	}
	return "OkToFly"
}

func rejectionResponse() (int, map[string]any) {
	return http.StatusOK, map[string]any{
		"activity_result":    "Rejected",
		"planning_result":    "Rejected",
		"flight_plan_status": "NotPlanned",
		// TODO(gap): Missing Fields: flight_id, includes_advisories, notes, queries(?), log_messages(?)
	}
}

func (h *Handler) hasAnyOtherIntent(entityID scdussv1.EntityID) bool {
	intents := slices.Collect(h.DB.GetAllIntents())
	return slices.IndexFunc(intents, func(intent db.OperationalIntent) bool {
		return intent.EntityID != entityID
	}) >= 0
}

const planningHorizon = 30 * 24 * time.Hour

func isTooEager(flight FlightPlan) bool {
	return time.Now().Add(planningHorizon).Before(flight.StartTime())
}

func hasEnded(flight FlightPlan) bool {
	return flight.EndTime().Before(time.Now())
}
