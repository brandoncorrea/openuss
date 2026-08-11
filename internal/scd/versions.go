package scd

import (
	"encoding/json"
	"net/http"
)

func GetVersion(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"system_identity": "astm.f3548.v21",
		"system_version":  "",
	})
}
