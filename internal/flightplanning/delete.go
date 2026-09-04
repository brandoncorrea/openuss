package flightplanning

import (
	"net/http"
	"uuid"

	"bwawan.com/openuss/internal/api"
)

func (handler *Handler) DeleteFlightPlan(w http.ResponseWriter, r *http.Request) {
	flightId, ok := parseUuid(r.PathValue("flight_plan_id"))
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	flight := handler.DB.GetFlight(flightId)
	if flight == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	intent := handler.DB.GetIntent(flight.EntityID)
	if intent != nil {
		_, err := handler.DSS.DeleteOperationalIntent(r.Context(), intent.EntityID, intent.Ovn)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		handler.DB.DeleteIntent(intent.EntityID)
	}
	handler.DB.DeleteFlight(flight.Id)
	api.WriteJSON(w, http.StatusOK, map[string]any{
		"flight_plan_status": "Closed",
		"planning_result":    "Completed",
	})
}

func parseUuid(s string) (uuid.UUID, bool) {
	if s == "" {
		return uuid.UUID{}, false
	}
	result, err := uuid.Parse(s)
	return result, err == nil
}
