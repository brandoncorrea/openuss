package scd

import (
	"context"
	"slices"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/internal/volume"
)

func isBlockingUs(
	other OperationalIntent,
	ownIntent OperationalIntent,
	params IntentParams,
) bool {
	if other.EntityID == ownIntent.EntityID {
		return false
	}
	// TODO(gap): This is the only schema validation we do against peer intents
	if other.State == "Flying" {
		return true
	}
	if params.Priority > other.Priority {
		return false
	}
	if !volume.VolumesIntersect(params.Volumes, other.Volumes) {
		return false
	}
	if ownIntent.State == scdussv1.OperationalIntentState_Activated {
		return !volume.VolumesIntersect(ownIntent.Volumes, other.Volumes)
	}
	return true
}

func hasKnownConflict(
	ownIntent OperationalIntent,
	params IntentParams,
	knownIntents []OperationalIntent,
) bool {
	return slices.ContainsFunc(knownIntents, func(other OperationalIntent) bool {
		return isBlockingUs(other, ownIntent, params)
	})
}

func (s *Service) resolvePeerConflict(
	ctx context.Context,
	missing scdussv1.OperationalIntentReference,
	ownIntent OperationalIntent,
	params IntentParams,
) (scdussv1.EntityOVN, error) {
	details, err := s.Peer.GetOperationalIntentDetails(ctx, missing.UssBaseUrl, missing.Id)
	// TODO(gap): We assume any peer error is a conflict
	if err != nil {
		return "", ErrConflict
	}
	peerIntent := intentFromDetails(details)
	if isBlockingUs(peerIntent, ownIntent, params) {
		return "", ErrConflict
	}
	return peerIntent.OVN, nil
}

// TODO(gap): Dereferences without nil checks
func intentFromDetails(details scdussv1.GetOperationalIntentDetailsResponse) OperationalIntent {
	return OperationalIntent{
		State:    details.OperationalIntent.Reference.State,
		Priority: *details.OperationalIntent.Details.Priority,
		Volumes:  *details.OperationalIntent.Details.Volumes,
		OVN:      *details.OperationalIntent.Reference.Ovn,
	}
}
