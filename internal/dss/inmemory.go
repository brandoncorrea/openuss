package dss

import (
	"context"
	"fmt"
	"uuid"

	"bwawan.com/openuss/internal/api/scdussv1"
)

type InMemoryDSS struct {
	Intents map[scdussv1.EntityID]scdussv1.OperationalIntent
}

func NewInMemoryDSS() *InMemoryDSS {
	return &InMemoryDSS{
		Intents: map[scdussv1.EntityID]scdussv1.OperationalIntent{},
	}
}

func (dss *InMemoryDSS) CreateOperationalIntentReference(
	ctx context.Context,
	id scdussv1.EntityID,
	params scdussv1.PutOperationalIntentReferenceParameters,
) (scdussv1.ChangeOperationalIntentReferenceResponse, error) {
	mustHaveContext(ctx)
	dss.Intents[id] = scdussv1.OperationalIntent{
		Reference: scdussv1.OperationalIntentReference{
			Id:              id,
			Manager:         "InMemoryManager",
			UssAvailability: scdussv1.UssAvailabilityState_Normal,
			Version:         1,
			State:           params.State,
			Ovn:             new(scdussv1.EntityOVN(uuid.New().String())),
			TimeStart: scdussv1.Time{
				Value:  params.Extents[0].TimeStart.Value,
				Format: "RFC3339",
			},
			TimeEnd: scdussv1.Time{
				Value:  params.Extents[0].TimeEnd.Value,
				Format: "RFC3339",
			},
			UssBaseUrl:     params.UssBaseUrl,
			SubscriptionId: scdussv1.SubscriptionID(uuid.New().String()),
		},
		Details: scdussv1.OperationalIntentDetails{
			Volumes: &params.Extents,
		},
	}
	return scdussv1.ChangeOperationalIntentReferenceResponse{
		OperationalIntentReference: dss.Intents[id].Reference,
	}, nil
}

func (dss *InMemoryDSS) DeleteOperationalIntent(
	ctx context.Context,
	id scdussv1.EntityID,
	ovn scdussv1.EntityOVN,
) (response scdussv1.ChangeOperationalIntentReferenceResponse, err error) {
	mustHaveContext(ctx)
	intent, ok := dss.Intents[id]
	if !ok {
		err = fmt.Errorf("dss: operational intent not found %v", id)
		return
	}
	if *intent.Reference.Ovn != ovn {
		err = fmt.Errorf("dss: supplied OVN does not match")
		return
	}
	delete(dss.Intents, id)
	return scdussv1.ChangeOperationalIntentReferenceResponse{}, nil
}

func mustHaveContext(ctx context.Context) {
	if ctx == nil {
		panic("inmemory dss: missing Context")
	}
}
