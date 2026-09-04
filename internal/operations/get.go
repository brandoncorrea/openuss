package operations

import (
	"net/http"
	"time"

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
		OperationalIntent: scdussv1.OperationalIntent{
			Reference: scdussv1.OperationalIntentReference{
				Id:              intent.EntityID,
				Manager:         intent.Manager,
				UssAvailability: intent.UssAvailability,
				Version:         intent.Version,
				State:           intent.State,
				Ovn:             &intent.Ovn,
				TimeStart: scdussv1.Time{
					Value:  intent.TimeStart.Format(time.RFC3339Nano),
					Format: "RFC3339",
				},
				TimeEnd: scdussv1.Time{
					Value:  intent.TimeEnd.Format(time.RFC3339Nano),
					Format: "RFC3339",
				},
				UssBaseUrl:     intent.UssBaseUrl,
				SubscriptionId: intent.SubscriptionId,
			},
			Details: scdussv1.OperationalIntentDetails{
				Volumes: &intent.Volumes,
			},
		},
	})
}
