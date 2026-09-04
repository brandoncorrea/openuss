package flightplanning

import (
	"net/http"

	"bwawan.com/openuss/internal/api"
	"bwawan.com/openuss/internal/api/scdussv1"
)

func (handler *Handler) DeleteFlightPlan(w http.ResponseWriter, r *http.Request) {
	flightId := r.PathValue("flight_plan_id")
	if flightId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	entityId := scdussv1.EntityID(flightId)
	intent := handler.DB.GetIntent(entityId)
	if intent == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_, err := handler.DSS.DeleteOperationalIntent(r.Context(), scdussv1.EntityID(flightId), scdussv1.EntityOVN(*intent.Reference.Ovn))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	handler.DB.DeleteIntent(intent.Reference.Id)
	api.WriteJSON(w, http.StatusOK, map[string]any{
		"flight_plan_status": "Closed",
		"planning_result":    "Completed",
	})
}
