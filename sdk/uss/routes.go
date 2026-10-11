package uss

import (
	"net/http"

	"bwawan.com/openuss/sdk/api"
	"bwawan.com/openuss/sdk/api/scdussv1"
)

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /uss/v1/operational_intents/{entity_id}",
		h.protect(scdussv1.GetOperationalIntentDetailsSecurity, h.GetOperationalIntent))
	mux.Handle("POST /uss/v1/operational_intents",
		h.protect(scdussv1.NotifyOperationalIntentDetailsChangedSecurity, h.NotifyOperationalIntent))
	mux.Handle("POST /uss/v1/reports",
		h.protect(scdussv1.MakeUssReportSecurity, h.MakeUSSReport))
}

func (h *Handler) protect(options []api.AuthorizationOption, next http.HandlerFunc) Protected {
	return Protected{
		Authorizer: h.Authorizer,
		Options:    options,
		Next:       next,
	}
}
