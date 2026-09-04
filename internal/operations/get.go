package operations

import (
	"net/http"

	"bwawan.com/openuss/internal/api"
	"bwawan.com/openuss/internal/api/scdussv1"
)

func (handler *Handler) GetOperationalIntent(w http.ResponseWriter, r *http.Request) {
	entityId := r.PathValue("entity_id")
	if entityId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	intent := handler.DB.GetIntent(scdussv1.EntityID(entityId))
	if intent == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	api.WriteJSON(w, http.StatusOK, scdussv1.GetOperationalIntentDetailsResponse{
		OperationalIntent: *intent,
	})
}
