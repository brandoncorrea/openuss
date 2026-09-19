package flightplanning

import (
	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/db"
	"bwawan.com/openuss/internal/dss"
	"bwawan.com/openuss/internal/peer"
)

type Handler struct {
	DSS        dss.USSAuthority
	Peer       peer.Client
	DB         db.DB
	UssBaseUrl scdussv1.OperationalIntentUssBaseURL
}

func New(
	dss dss.USSAuthority,
	peer peer.Client,
	db db.DB,
	baseUrl string,
) *Handler {
	return &Handler{
		DSS:        dss,
		Peer:       peer,
		DB:         db,
		UssBaseUrl: scdussv1.OperationalIntentUssBaseURL(baseUrl),
	}
}
