package flightplanning

import (
	"net/http"
	"uuid"

	"bwawan.com/openuss/internal/api"
)

func (h *Handler) DeleteFlightPlan(w http.ResponseWriter, r *http.Request) {
	flightID, ok := parseUUID(r.PathValue("flight_plan_id"))
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	flight := h.DB.GetFlight(flightID)
	if flight == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	intent := h.DB.GetIntent(flight.EntityID)
	if intent != nil {
		_, err := h.DSS.DeleteOperationalIntent(r.Context(), intent.EntityID, intent.OVN)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.DB.DeleteIntent(intent.EntityID)
	}
	h.DB.DeleteFlight(flight.ID)
	api.WriteJSON(w, http.StatusOK, map[string]any{
		"flight_plan_status": "Closed",
		"planning_result":    "Completed",
	})
}

func parseUUID(s string) (uuid.UUID, bool) {
	if s == "" {
		return uuid.UUID{}, false
	}
	result, err := uuid.Parse(s)
	return result, err == nil
}
