package flightplanning

import (
	"net/http"

	"bwawan.com/openuss/internal/api"
)

func GetStatus(w http.ResponseWriter, _ *http.Request) {
	api.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "Ready",
	})
}
