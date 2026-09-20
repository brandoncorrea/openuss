package uss

import "net/http"

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /uss/v1/operational_intents/{entity_id}", h.GetOperationalIntent)
}
