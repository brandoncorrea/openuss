package flightplanning

import (
	"net/http"

	"bwawan.com/openuss/sdk/api"
)

func (*Handler) ClearAreaRequests(w http.ResponseWriter, _ *http.Request) {
	// TODO(gap): Reports success without clearing anything
	api.WriteJSON(w, http.StatusOK, ClearAreaResponse{
		Outcome: ClearAreaOutcome{
			Success: true,
		},
	})
}
