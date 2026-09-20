package operations

import (
	"errors"
	"net/http"
	"time"

	"bwawan.com/openuss/sdk/api"
	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/scd"
)

func (h *Handler) GetOperationalIntent(w http.ResponseWriter, r *http.Request) {
	entityID := r.PathValue("entity_id")
	if entityID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// TODO: Missing context
	intent, err := h.Intents.Get(nil, scdussv1.EntityID(entityID))

	// TODO: What if there is a different kind of error?
	if errors.Is(err, scd.ErrNotFound) {
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
