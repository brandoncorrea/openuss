package versioning

import (
	"net/http"

	"bwawan.com/openuss/internal/util"
)

func GetVersion(w http.ResponseWriter, _ *http.Request) {
	util.WriteJSON(w, map[string]string{
		"system_identity": "astm.f3548.v21",
		"system_version":  "",
	})
}
