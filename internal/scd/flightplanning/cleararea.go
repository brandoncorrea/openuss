package flightplanning

import (
	"net/http"

	"bwawan.com/openuss/internal/api"
)

func ClearArea(w http.ResponseWriter, _ *http.Request) {
	api.WriteJSON(w, http.StatusOK, map[string]any{
		"outcome": map[string]any{
			"success": true,
		},
	})
}
