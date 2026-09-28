package scd

import (
	"context"
	"errors"
	"time"
	"uuid"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/dss"
	"bwawan.com/openuss/sdk/internal/util"
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
	// TODO: Missing context
	// TODO(gap): No error handling
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
	if err := params.validate(); err != nil {
		return OperationalIntent{}, err
	}
	if params.State == scdussv1.OperationalIntentState_Activated && ovn == nil {
		return OperationalIntent{}, ErrConflict
	}

	// TODO: Missing context
	// TODO(gap): No error handling
	// TODO(gap): This includes our own intent on Update
	knownIntents, _ := s.Intents.List(nil)
	if hasKnownConflict(intent, params, knownIntents) {
		return OperationalIntent{}, ErrRejected
	}

	result, err := s.submitOperationalIntent(ctx, intent, ovn, params, knownIntents)
	if err != nil {
		return OperationalIntent{}, err
	}

	s.notifyPeers(ctx, result)
	return s.saveOperationalIntent(params, result), nil
}

func (s *Service) submitOperationalIntent(
	ctx context.Context,
	intent OperationalIntent,
	ovn *scdussv1.EntityOVN,
	params IntentParams,
	knownIntents []OperationalIntent,
) (scdussv1.ChangeOperationalIntentReferenceResponse, error) {
	putParams := s.createPutRequestParams(params, keyFromIntents(knownIntents))

	// TODO(gap): What happens if the DSS call results in a non-conflict error?
	result, err := s.DSS.PutOperationalIntentReference(ctx, intent.EntityID, ovn, putParams)
	if conflict, ok := errors.AsType[dss.AirspaceConflictError](err); ok {
		// TODO(gap): We only fetch the first missing intent - the rest are ignored
		missing := (*conflict.MissingOperationalIntents)[0]
		peerOVN, err := s.resolvePeerConflict(ctx, missing, intent, params)
		if err != nil {
			return scdussv1.ChangeOperationalIntentReferenceResponse{}, err
		}
		putParams.Key = new(append(*putParams.Key, peerOVN))
		// TODO(gap): We assume the DSS won't return a subsequent conflict
		result, _ = s.DSS.PutOperationalIntentReference(ctx, intent.EntityID, ovn, putParams)
	}
	return result, nil
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
			UssBaseUrl: scdussv1.SubscriptionUssBaseURL(s.USSBaseURL),
		},
	}
}

// TODO(gap): This sends ALL OVNs, not just the relevant ones
func keyFromIntents(intents []OperationalIntent) scdussv1.Key {
	return util.Map(intents, func(intent OperationalIntent) scdussv1.EntityOVN {
		return intent.OVN
	})
}

// TODO(gap): Nothing validates the DSS response
func (s *Service) saveOperationalIntent(
	params IntentParams,
	result scdussv1.ChangeOperationalIntentReferenceResponse,
) OperationalIntent {
	// TODO(gap): Unparseable times are zero valued
	timeStart, _ := time.Parse(time.RFC3339Nano, result.OperationalIntentReference.TimeStart.Value)
	timeEnd, _ := time.Parse(time.RFC3339Nano, result.OperationalIntentReference.TimeEnd.Value)
	intent := OperationalIntent{
		EntityID:        result.OperationalIntentReference.Id,
		Manager:         result.OperationalIntentReference.Manager,
		USSAvailability: result.OperationalIntentReference.UssAvailability,
		Version:         result.OperationalIntentReference.Version,
		Priority:        params.Priority,
		State:           result.OperationalIntentReference.State,
		// TODO(gap): Missing OVN panics
		OVN:            *result.OperationalIntentReference.Ovn,
		TimeStart:      timeStart,
		TimeEnd:        timeEnd,
		USSBaseURL:     result.OperationalIntentReference.UssBaseUrl,
		SubscriptionID: result.OperationalIntentReference.SubscriptionId,
		Volumes:        params.Volumes,
	}
	// TODO: Missing context
	// TODO(gap): No error handling
	s.Intents.Upsert(nil, intent)
	return intent
}
