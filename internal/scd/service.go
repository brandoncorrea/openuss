package scd

import (
	"bwawan.com/openuss/internal/dss"
	"bwawan.com/openuss/internal/peer"
	"bwawan.com/openuss/sdk/api/scdussv1"
)

type Service struct {
	DSS        dss.Client
	Peer       peer.Client
	Intents    IntentStore
	USSBaseURL scdussv1.OperationalIntentUssBaseURL
}

func New(
	dssClient dss.Client,
	peerClient peer.Client,
	intents IntentStore,
	baseURL string,
) *Service {
	return &Service{
		DSS:        dssClient,
		Peer:       peerClient,
		Intents:    intents,
		USSBaseURL: scdussv1.OperationalIntentUssBaseURL(baseURL),
	}
}
