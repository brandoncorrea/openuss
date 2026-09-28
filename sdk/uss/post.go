package uss

import "net/http"

// TODO(gap): Currently does nothing
func (h *Handler) NotifyOperationalIntent(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
