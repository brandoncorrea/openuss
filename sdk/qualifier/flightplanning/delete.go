package flightplanning

import (
	"net/http"
	"uuid"

	"bwawan.com/openuss/sdk/api"
)

func (h *Handler) DeleteFlightPlan(w http.ResponseWriter, r *http.Request) {
	flightID, ok := parseUUID(r.PathValue("flight_plan_id"))
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	flight := h.Flights.Get(flightID)
	if flight == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := h.SCD.DeleteOperationalIntent(r.Context(), flight.EntityID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.Flights.Delete(flight.ID)
	api.WriteJSON(w, http.StatusOK, FlightPlanResponse{
		FlightPlanStatus: FlightPlanStatusClosed,
		PlanningResult:   PlanningActivityResultCompleted,
	})
}

func parseUUID(s string) (uuid.UUID, bool) {
	if s == "" {
		return uuid.UUID{}, false
	}
	result, err := uuid.Parse(s)
	return result, err == nil
}
