package uss

import (
	"net/http"
	"time"

	"bwawan.com/openuss/sdk/api"
	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/scd"
)

func (h *Handler) GetOperationalIntent(w http.ResponseWriter, r *http.Request) {
	// TODO: Missing context
	// TODO(gap): No error handling; always 200
	intent, _ := h.Intents.Get(nil, scdussv1.EntityID(r.PathValue("entity_id")))
	api.WriteJSON(w, http.StatusOK, toWire(intent))
}

func toWire(intent scd.OperationalIntent) scdussv1.GetOperationalIntentDetailsResponse {
	return scdussv1.GetOperationalIntentDetailsResponse{
		OperationalIntent: scdussv1.OperationalIntent{
			Reference: scdussv1.OperationalIntentReference{
				Id:              intent.EntityID,
				Manager:         intent.Manager,
				UssAvailability: intent.USSAvailability,
				Version:         intent.Version,
				State:           intent.State,
				Ovn:             &intent.OVN,
				TimeStart:       toSCDTime(intent.TimeStart),
				TimeEnd:         toSCDTime(intent.TimeEnd),
				UssBaseUrl:      intent.USSBaseURL,
				SubscriptionId:  intent.SubscriptionID,
			},
			Details: scdussv1.OperationalIntentDetails{
				Volumes:  &intent.Volumes,
				Priority: &intent.Priority,
			},
		},
	}
}

func toSCDTime(value time.Time) scdussv1.Time {
	return scdussv1.Time{
		Value:  value.Format(time.RFC3339Nano),
		Format: "RFC3339",
	}
}
