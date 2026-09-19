package peer

import (
	"context"
	"fmt"

	"bwawan.com/openuss/internal/api/scdussv1"
)

type InMemoryPeer struct {
	References []scdussv1.OperationalIntentReference
}

func NewInMemoryPeer(peers []scdussv1.OperationalIntentReference) *InMemoryPeer {
	return &InMemoryPeer{References: peers}
}

func (p *InMemoryPeer) GetOperationalIntentDetails(
	_ context.Context,
	ussBaseURL scdussv1.OperationalIntentUssBaseURL,
	entityID scdussv1.EntityID,
) (scdussv1.GetOperationalIntentDetailsResponse, error) {
	for _, intent := range p.References {
		if intent.Id == entityID && intent.UssBaseUrl == ussBaseURL {
			return scdussv1.GetOperationalIntentDetailsResponse{
				OperationalIntent: scdussv1.OperationalIntent{
					Reference: intent,
				},
			}, nil
		}
	}
	err := fmt.Errorf("operational intent not found on %s: %s", ussBaseURL, entityID)
	return scdussv1.GetOperationalIntentDetailsResponse{}, err
}
