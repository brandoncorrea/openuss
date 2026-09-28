package flightplanning

import (
	"net/http"
	"uuid"

	"bwawan.com/openuss/sdk/api"
)

func (h *Handler) DeleteFlightPlan(w http.ResponseWriter, r *http.Request) {
	// TODO(gap): What if uuid parse fails?
	flightID, _ := uuid.Parse(r.PathValue("flight_plan_id"))
	flight := h.Flights.Get(flightID)

	// TODO(gap): 'flight' is used without a nil check
	// TODO(gap): No error checks - assumes success
	h.SCD.DeleteOperationalIntent(r.Context(), flight.EntityID)
	h.Flights.Delete(flight.ID)

	api.WriteJSON(w, http.StatusOK, FlightPlanResponse{
		FlightPlanStatus: FlightPlanStatusClosed,
		PlanningResult:   PlanningActivityResultCompleted,
	})
}
