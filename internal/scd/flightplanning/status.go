package flightplanning

import (
	"net/http"

	"bwawan.com/openuss/internal/util"
)

func GetStatus(w http.ResponseWriter, _ *http.Request) {
	util.WriteJSON(w, map[string]string{
		"status": "Ready",
	})
}
