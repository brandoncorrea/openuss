package uss

import (
	"bwawan.com/openuss/sdk/api"
	"bwawan.com/openuss/sdk/scd"
)

type Handler struct {
	Intents    scd.IntentStore
	Authorizer api.Authorizer
}

func New(intents scd.IntentStore, authorizer api.Authorizer) *Handler {
	return &Handler{
		Intents:    intents,
		Authorizer: authorizer,
	}
}
