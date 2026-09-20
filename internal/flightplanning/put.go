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
	"bwawan.com/openuss/internal/scd"
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
	existingFlight := h.DB.GetFlight(flightID)
	entityID, ovn := h.findIDsForExistingFlightPlan(existingFlight)

	// TODO(gap): Validate flight_plan_id is a valid UUID, among other things
	intent, err := h.putIntent(ctx, entityID, ovn, toIntentParams(plan))
	if errors.Is(err, scd.ErrConflict) {
		return conflictResponse(ovn)
	}
	if errors.Is(err, scd.ErrRejected) {
		return rejectionResponse()
	}

	h.DB.SaveFlight(db.FlightPlan{
		ID:       flightID,
		EntityID: intent.EntityID,
	})

	return http.StatusOK, map[string]any{
		"planning_result":    "Completed",
		"flight_plan_status": flightPlanStatus(ovn),
		// TODO(gap): Missing Fields: activity_result, as_planned, flight_id, includes_advisories, queries(?), log_messages(?)
	}
}

func (h *Handler) putIntent(
	ctx context.Context,
	entityID scdussv1.EntityID,
	ovn *scdussv1.EntityOVN,
	params scd.IntentParams,
) (scd.OperationalIntent, error) {
	if isInvalidIntent(params) {
		return scd.OperationalIntent{}, scd.ErrRejected
	}

	// TODO(gap): This assumes everything overlaps
	if h.hasAnyOtherIntent(entityID) {
		return scd.OperationalIntent{}, scd.ErrRejected
	}

	putParams := h.createPutRequestParams(params)

	// TODO(gap): What happens if the DSS call results in a non-conflict error?
	result, err := h.DSS.PutOperationalIntentReference(ctx, entityID, ovn, putParams)
	if conflict, ok := errors.AsType[dss.AirspaceConflictError](err); ok {
		missing := (*conflict.MissingOperationalIntents)[0]
		details, _ := h.Peer.GetOperationalIntentDetails(ctx, missing.UssBaseUrl, missing.Id)
		if isLowerPriority(details) {
			return scd.OperationalIntent{}, scd.ErrConflict
		}
		putParams.Key = &scdussv1.Key{*details.OperationalIntent.Reference.Ovn}
		result, err = h.DSS.PutOperationalIntentReference(ctx, entityID, ovn, putParams)
	}

	return h.saveOperationalIntent(params, result), nil
}

func isLowerPriority(details scdussv1.GetOperationalIntentDetailsResponse) bool {
	return details.OperationalIntent.Details.Priority != nil &&
		*details.OperationalIntent.Details.Priority == 100 &&
		details.OperationalIntent.Reference.State != scdussv1.OperationalIntentState_Activated
}

func toIntentParams(plan FlightPlan) scd.IntentParams {
	return scd.IntentParams{
		Volumes:  plan.BasicInformation.Area,
		State:    flightPlanState(plan),
		Priority: scdussv1.Priority(plan.F3548.Priority),
	}
}

func isInvalidIntent(params scd.IntentParams) bool {
	return isTooEager(params) || hasEnded(params)
}

func (h *Handler) findIDsForExistingFlightPlan(flight *db.FlightPlan) (scdussv1.EntityID, *scdussv1.EntityOVN) {
	if flight == nil {
		return scdussv1.EntityID(uuid.New().String()), nil
	}
	// TODO: Missing context; no error handling
	intent, _ := h.Intents.Get(nil, flight.EntityID)
	return intent.EntityID, new(intent.OVN)
}

func (h *Handler) createPutRequestParams(
	params scd.IntentParams,
) scdussv1.PutOperationalIntentReferenceParameters {
	return scdussv1.PutOperationalIntentReferenceParameters{
		Extents:    params.Volumes,
		State:      params.State,
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
	params scd.IntentParams,
	result scdussv1.ChangeOperationalIntentReferenceResponse,
) scd.OperationalIntent {
	timeStart, _ := time.Parse(time.RFC3339Nano, result.OperationalIntentReference.TimeStart.Value)
	timeEnd, _ := time.Parse(time.RFC3339Nano, result.OperationalIntentReference.TimeEnd.Value)
	intent := scd.OperationalIntent{
		EntityID:        result.OperationalIntentReference.Id,
		Manager:         result.OperationalIntentReference.Manager,
		USSAvailability: result.OperationalIntentReference.UssAvailability,
		Version:         result.OperationalIntentReference.Version,
		Priority:        params.Priority,
		State:           result.OperationalIntentReference.State,
		OVN:             *result.OperationalIntentReference.Ovn,
		TimeStart:       timeStart,
		TimeEnd:         timeEnd,
		USSBaseURL:      result.OperationalIntentReference.UssBaseUrl,
		SubscriptionID:  result.OperationalIntentReference.SubscriptionId,
		Volumes:         params.Volumes,
	}
	// TODO: Missing context; no error handling
	h.Intents.Upsert(nil, intent)
	return intent
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

func conflictResponse(ovn *scdussv1.EntityOVN) (int, map[string]any) {
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

func (h *Handler) hasAnyOtherIntent(entityID scdussv1.EntityID) bool {
	// TODO: Missing context; no error handling
	intents, _ := h.Intents.List(nil)
	return slices.IndexFunc(intents, func(intent scd.OperationalIntent) bool {
		return intent.EntityID != entityID
	}) >= 0
}

const planningHorizon = 30 * 24 * time.Hour

func isTooEager(params scd.IntentParams) bool {
	return time.Now().Add(planningHorizon).Before(startTime(params))
}

func hasEnded(params scd.IntentParams) bool {
	return endTime(params).Before(time.Now())
}

func startTime(params scd.IntentParams) time.Time {
	// TODO(gap): Nothing validates a zero-area or multi-area flight plan
	return rfc3339(params.Volumes[0].TimeStart.Value)
}

func endTime(params scd.IntentParams) time.Time {
	// TODO(gap): Nothing validates a zero-area or multi-area flight plan
	return rfc3339(params.Volumes[0].TimeEnd.Value)
}

func rfc3339(s string) time.Time {
	// TODO(gap): Nothing validates a malformed timestamp
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
