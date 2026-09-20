package flightplanning

import (
	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/db"
	"bwawan.com/openuss/internal/dss"
	"bwawan.com/openuss/internal/peer"
	"bwawan.com/openuss/internal/scd"
)

type Handler struct {
	DSS        dss.Client
	Peer       peer.Client
	DB         db.DB
	Intents    scd.IntentStore
	USSBaseURL scdussv1.OperationalIntentUssBaseURL
	SCD        *scd.Service
}

func New(
	dssClient dss.Client,
	service *scd.Service,
	peer peer.Client,
	db db.DB,
	intents scd.IntentStore,
	baseURL string,
) *Handler {
	return &Handler{
		DSS:        dssClient,
		SCD:        service,
		Peer:       peer,
		DB:         db,
		Intents:    intents,
		USSBaseURL: scdussv1.OperationalIntentUssBaseURL(baseURL),
	}
}
