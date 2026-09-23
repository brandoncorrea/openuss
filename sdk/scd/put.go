package scd

import (
	"context"
	"errors"
	"slices"
	"time"
	"uuid"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/dss"
	"bwawan.com/openuss/sdk/internal/volume"
)

func (s *Service) CreateOperationalIntent(
	ctx context.Context,
	params IntentParams,
) (OperationalIntent, error) {
	id := scdussv1.EntityID(uuid.New().String())
	intent := OperationalIntent{EntityID: id}
	return s.put(ctx, intent, nil, params)
}

func (s *Service) UpdateOperationalIntent(
	ctx context.Context,
	id scdussv1.EntityID,
	params IntentParams,
) (OperationalIntent, error) {
	// TODO: Missing context; no error handling
	existing, _ := s.Intents.Get(nil, id)
	updated, err := s.put(ctx, existing, new(existing.OVN), params)
	if err != nil {
		return existing, err
	}
	return updated, nil
}

func (s *Service) put(
	ctx context.Context,
	intent OperationalIntent,
	ovn *scdussv1.EntityOVN,
	params IntentParams,
) (OperationalIntent, error) {
	if isInvalidIntent(params) {
		return OperationalIntent{}, ErrRejected
	}

	// TODO(gap): This assumes everything overlaps
	if s.hasAnyOtherIntent(intent.EntityID) {
		return OperationalIntent{}, ErrRejected
	}

	putParams := s.createPutRequestParams(params)

	// TODO(gap): What happens if the DSS call results in a non-conflict error?
	result, err := s.DSS.PutOperationalIntentReference(ctx, intent.EntityID, ovn, putParams)
	if conflict, ok := errors.AsType[dss.AirspaceConflictError](err); ok {
		missing := (*conflict.MissingOperationalIntents)[0]
		details, _ := s.Peer.GetOperationalIntentDetails(ctx, missing.UssBaseUrl, missing.Id)
		peerIntent := detailResponseToIntent(details)
		if peerBlocksUs(peerIntent, intent, params) {
			return OperationalIntent{}, ErrConflict
		}
		putParams.Key = &scdussv1.Key{*details.OperationalIntent.Reference.Ovn}
		result, _ = s.DSS.PutOperationalIntentReference(ctx, intent.EntityID, ovn, putParams)
	}

	return s.saveOperationalIntent(params, result), nil
}

const blockingPriority = 100

func peerBlocksUs(
	peer OperationalIntent,
	intent OperationalIntent,
	params IntentParams,
) bool {
	if !volume.VolumesIntersect(params.Volumes, peer.Volumes) {
		return false
	}
	if peer.Priority != blockingPriority {
		return false
	}
	if intent.State == scdussv1.OperationalIntentState_Activated {
		return !volume.VolumesIntersect(intent.Volumes, peer.Volumes)
	}
	return true
}

func isInvalidIntent(params IntentParams) bool {
	return isTooEager(params) || hasEnded(params)
}

func (s *Service) createPutRequestParams(
	params IntentParams,
) scdussv1.PutOperationalIntentReferenceParameters {
	return scdussv1.PutOperationalIntentReferenceParameters{
		Extents:    params.Volumes,
		State:      params.State,
		UssBaseUrl: s.USSBaseURL,
		NewSubscription: &scdussv1.ImplicitSubscriptionParameters{
			// TODO(gap): This probably needs to be a proper URL
			UssBaseUrl: scdussv1.SubscriptionUssBaseURL("x"),
		},
	}
}

func (s *Service) saveOperationalIntent(
	params IntentParams,
	result scdussv1.ChangeOperationalIntentReferenceResponse,
) OperationalIntent {
	timeStart, _ := time.Parse(time.RFC3339Nano, result.OperationalIntentReference.TimeStart.Value)
	timeEnd, _ := time.Parse(time.RFC3339Nano, result.OperationalIntentReference.TimeEnd.Value)
	intent := OperationalIntent{
		EntityID:        result.OperationalIntentReference.Id,
		Manager:         result.OperationalIntentReference.Manager,
		USSAvailability: result.OperationalIntentReference.UssAvailability,
		Version:         result.OperationalIntentReference.Version,
		Priority:        params.Priority,
		State:           result.OperationalIntentReference.State,
		OVN:             *result.OperationalIntentReference.Ovn,
		TimeStart:       timeStart,
		TimeEnd:         timeEnd,
		USSBaseURL:      result.OperationalIntentReference.UssBaseUrl,
		SubscriptionID:  result.OperationalIntentReference.SubscriptionId,
		Volumes:         params.Volumes,
	}
	// TODO: Missing context; no error handling
	s.Intents.Upsert(nil, intent)
	return intent
}

func (s *Service) hasAnyOtherIntent(entityID scdussv1.EntityID) bool {
	// TODO: Missing context; no error handling
	intents, _ := s.Intents.List(nil)
	return slices.IndexFunc(intents, func(intent OperationalIntent) bool {
		return intent.EntityID != entityID
	}) >= 0
}

func detailResponseToIntent(details scdussv1.GetOperationalIntentDetailsResponse) OperationalIntent {
	return OperationalIntent{
		Priority: *details.OperationalIntent.Details.Priority,
		Volumes:  *details.OperationalIntent.Details.Volumes,
		OVN:      *details.OperationalIntent.Reference.Ovn,
	}
}

const planningHorizon = 30 * 24 * time.Hour

func isTooEager(params IntentParams) bool {
	return time.Now().Add(planningHorizon).Before(startTime(params))
}

func hasEnded(params IntentParams) bool {
	return endTime(params).Before(time.Now())
}

func startTime(params IntentParams) time.Time {
	// TODO(gap): Nothing validates a zero-area or multi-area flight plan
	return rfc3339(params.Volumes[0].TimeStart.Value)
}

func endTime(params IntentParams) time.Time {
	// TODO(gap): Nothing validates a zero-area or multi-area flight plan
	return rfc3339(params.Volumes[0].TimeEnd.Value)
}

func rfc3339(s string) time.Time {
	// TODO(gap): Nothing validates a malformed timestamp
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
