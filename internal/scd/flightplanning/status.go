package flightplanning

import (
	"net/http"

	"bwawan.com/openuss/internal/api"
)

func GetStatus(w http.ResponseWriter, _ *http.Request) {
	// TODO: Probably need to actually check a "Ready" state
	api.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "Ready",
	})
}
