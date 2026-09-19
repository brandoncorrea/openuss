package operations

import (
	"net/http"
	"time"

	"bwawan.com/openuss/internal/api"
	"bwawan.com/openuss/internal/api/scdussv1"
)

func (h *Handler) GetOperationalIntent(w http.ResponseWriter, r *http.Request) {
	entityID := r.PathValue("entity_id")
	if entityID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	intent := h.DB.GetIntent(scdussv1.EntityID(entityID))
	if intent == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	api.WriteJSON(w, http.StatusOK, scdussv1.GetOperationalIntentDetailsResponse{
		OperationalIntent: scdussv1.OperationalIntent{
			Reference: scdussv1.OperationalIntentReference{
				Id:              intent.EntityID,
				Manager:         intent.Manager,
				UssAvailability: intent.USSAvailability,
				Version:         intent.Version,
				State:           intent.State,
				Ovn:             &intent.OVN,
				TimeStart: scdussv1.Time{
					Value:  intent.TimeStart.Format(time.RFC3339Nano),
					Format: "RFC3339",
				},
				TimeEnd: scdussv1.Time{
					Value:  intent.TimeEnd.Format(time.RFC3339Nano),
					Format: "RFC3339",
				},
				UssBaseUrl:     intent.USSBaseURL,
				SubscriptionId: intent.SubscriptionID,
			},
			Details: scdussv1.OperationalIntentDetails{
				Volumes:  &intent.Volumes,
				Priority: &intent.Priority,
			},
		},
	})
}
