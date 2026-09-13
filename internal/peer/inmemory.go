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

func (peer *InMemoryPeer) GetOperationalIntentDetails(
	_ context.Context,
	ussBaseUrl scdussv1.OperationalIntentUssBaseURL,
	entityId scdussv1.EntityID,
) (scdussv1.GetOperationalIntentDetailsResponse, error) {
	for _, intent := range peer.References {
		if intent.Id == entityId && intent.UssBaseUrl == ussBaseUrl {
			return scdussv1.GetOperationalIntentDetailsResponse{
				OperationalIntent: scdussv1.OperationalIntent{
					Reference: intent,
				},
			}, nil
		}
	}
	err := fmt.Errorf("operational intent not found on %s: %s", ussBaseUrl, entityId)
	return scdussv1.GetOperationalIntentDetailsResponse{}, err
}
