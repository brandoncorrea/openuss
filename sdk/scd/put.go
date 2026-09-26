package scd

import (
	"context"
	"errors"
	"slices"
	"time"
	"uuid"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/dss"
	"bwawan.com/openuss/sdk/internal/util"
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
	if params.State == scdussv1.OperationalIntentState_Nonconforming {
		return OperationalIntent{}, ErrNotSupported
	}
	if isInvalidIntent(params) {
		return OperationalIntent{}, ErrRejected
	}

	// TODO: Missing context; no error handling
	// TODO(gap): This includes our own intent on Update
	knownIntents, _ := s.Intents.List(nil)

	if hasKnownConflict(intent, params, knownIntents) {
		return OperationalIntent{}, ErrRejected
	}

	putParams := s.createPutRequestParams(params, keyFromIntents(knownIntents))

	if putParams.State == scdussv1.OperationalIntentState_Activated && ovn == nil {
		return OperationalIntent{}, ErrConflict
	}

	// TODO(gap): What happens if the DSS call results in a non-conflict error?
	result, err := s.DSS.PutOperationalIntentReference(ctx, intent.EntityID, ovn, putParams)
	if conflict, ok := errors.AsType[dss.AirspaceConflictError](err); ok {
		// TODO(gap): We only fetch the first missing intent - the rest are ignored
		missing := (*conflict.MissingOperationalIntents)[0]
		details, _ := s.Peer.GetOperationalIntentDetails(ctx, missing.UssBaseUrl, missing.Id)
		peerIntent := intentFromDetails(details)
		if isBlockingUs(peerIntent, intent, params) {
			return OperationalIntent{}, ErrConflict
		}
		putParams.Key = new(append(*putParams.Key, peerIntent.OVN))
		result, _ = s.DSS.PutOperationalIntentReference(ctx, intent.EntityID, ovn, putParams)
	}

	return s.saveOperationalIntent(params, result), nil
}

func isBlockingUs(
	other OperationalIntent,
	intent OperationalIntent,
	params IntentParams,
) bool {
	if params.Priority > other.Priority {
		return false
	}
	if !volume.VolumesIntersect(params.Volumes, other.Volumes) {
		return false
	}
	if intent.State == scdussv1.OperationalIntentState_Activated {
		return !volume.VolumesIntersect(intent.Volumes, other.Volumes)
	}
	return true
}

func isInvalidIntent(params IntentParams) bool {
	return isTooEager(params) || hasEnded(params)
}

func (s *Service) createPutRequestParams(
	params IntentParams,
	key scdussv1.Key,
) scdussv1.PutOperationalIntentReferenceParameters {
	return scdussv1.PutOperationalIntentReferenceParameters{
		Extents:    params.Volumes,
		State:      params.State,
		UssBaseUrl: s.USSBaseURL,
		Key:        new(key),
		NewSubscription: &scdussv1.ImplicitSubscriptionParameters{
			// TODO(gap): This probably needs to be a proper URL
			UssBaseUrl: scdussv1.SubscriptionUssBaseURL("x"),
		},
	}
}

// TODO(gap): This sends ALL OVNs, not just the relevant ones
func keyFromIntents(intents []OperationalIntent) scdussv1.Key {
	return util.Map(intents, func(intent OperationalIntent) scdussv1.EntityOVN {
		return intent.OVN
	})
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

func hasKnownConflict(
	intent OperationalIntent,
	params IntentParams,
	knownIntents []OperationalIntent,
) bool {
	return slices.IndexFunc(knownIntents, func(other OperationalIntent) bool {
		return other.EntityID != intent.EntityID && isBlockingUs(other, intent, params)
	}) >= 0
}

// TODO(gap): Dereferences without nil checks
func intentFromDetails(details scdussv1.GetOperationalIntentDetailsResponse) OperationalIntent {
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
