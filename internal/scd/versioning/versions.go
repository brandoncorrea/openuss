package versioning

import (
	"net/http"

	"bwawan.com/openuss/internal/api"
)

func GetVersion(w http.ResponseWriter, _ *http.Request) {
	api.WriteJSON(w, http.StatusOK, map[string]string{
		"system_identity": "astm.f3548.v21",
		"system_version":  "",
	})
}
