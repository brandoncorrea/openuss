package flightplanning

import (
	"net/http"

	"bwawan.com/openuss/internal/util"
)

func ClearArea(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	util.WriteJSON(w, map[string]any{
		"outcome": map[string]any{
			"success": true,
		},
	})
}
