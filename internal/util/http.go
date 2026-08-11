package util

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(body)
}
