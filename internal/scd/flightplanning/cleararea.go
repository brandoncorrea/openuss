package flightplanning

import (
	"net/http"

	"bwawan.com/openuss/internal/api"
)

func ClearArea(w http.ResponseWriter, _ *http.Request) {
	// TODO: Will likely need to actually clear the area
	api.WriteJSON(w, http.StatusOK, map[string]any{
		"outcome": map[string]any{
			"success": true,
		},
	})
}
