package router

import (
	"encoding/json"
	"net/http"
)

// New returns a handler covering every route OpenUSS serves.
func New() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /versioning/versions/astm.f3548.v21", GetVersion)
	return mux
}

func GetVersion(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"system_identity": "astm.f3548.v21",
		"system_version":  "",
	})
}
