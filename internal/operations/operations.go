package operations

import "bwawan.com/openuss/sdk/scd"

type Handler struct {
	Intents scd.IntentStore
}

func New(intents scd.IntentStore) *Handler {
	return &Handler{Intents: intents}
}
