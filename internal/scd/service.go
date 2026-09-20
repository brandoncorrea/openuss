package scd

import "bwawan.com/openuss/internal/dss"

type Service struct {
	DSS     dss.Client
	Intents IntentStore
}

func New(dssClient dss.Client, intents IntentStore) *Service {
	return &Service{
		DSS:     dssClient,
		Intents: intents,
	}
}
