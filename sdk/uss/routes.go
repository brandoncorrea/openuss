package uss

import "net/http"

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /uss/v1/operational_intents/{entity_id}", h.GetOperationalIntent)
	mux.HandleFunc("POST /uss/v1/operational_intents", h.NotifyOperationalIntent)
	mux.HandleFunc("POST /uss/v1/reports", h.MakeUSSReport)
}
