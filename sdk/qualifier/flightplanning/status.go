package flightplanning

import (
	"net/http"

	"bwawan.com/openuss/sdk/api"
)

func (*Handler) GetStatus(w http.ResponseWriter, _ *http.Request) {
	// TODO(gap): Always reports Ready
	api.WriteJSON(w, http.StatusOK, StatusResponse{
		Status: ServiceStatusReady,
	})
}
