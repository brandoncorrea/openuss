package dss

import (
	"context"
	"fmt"
	"uuid"

	"bwawan.com/openuss/internal/api/scdussv1"
)

type InMemoryDSS struct {
	Intents       map[scdussv1.EntityID]scdussv1.OperationalIntent
	Subscriptions map[scdussv1.SubscriptionID]scdussv1.Subscription
}

func NewInMemoryDSS() *InMemoryDSS {
	return &InMemoryDSS{
		Intents:       map[scdussv1.EntityID]scdussv1.OperationalIntent{},
		Subscriptions: map[scdussv1.SubscriptionID]scdussv1.Subscription{},
	}
}

func (d *InMemoryDSS) PutOperationalIntentReference(
	ctx context.Context,
	id scdussv1.EntityID,
	ovn *scdussv1.EntityOVN,
	params scdussv1.PutOperationalIntentReferenceParameters,
) (scdussv1.ChangeOperationalIntentReferenceResponse, error) {
	mustHaveContext(ctx)
	return scdussv1.ChangeOperationalIntentReferenceResponse{
		OperationalIntentReference: d.createOperationalIntentReference(id, ovn, params),
	}, nil
}

func (d *InMemoryDSS) createOperationalIntentReference(
	id scdussv1.EntityID,
	ovn *scdussv1.EntityOVN,
	params scdussv1.PutOperationalIntentReferenceParameters,
) scdussv1.OperationalIntentReference {
	if ovn == nil {
		ovn = new(scdussv1.EntityOVN(uuid.New().String()))
	}
	intent := scdussv1.OperationalIntent{
		Reference: scdussv1.OperationalIntentReference{
			Id:              id,
			Manager:         "InMemoryManager",
			UssAvailability: scdussv1.UssAvailabilityState_Normal,
			Version:         1,
			State:           params.State,
			Ovn:             ovn,
			TimeStart: scdussv1.Time{
				Value:  params.Extents[0].TimeStart.Value,
				Format: "RFC3339",
			},
			TimeEnd: scdussv1.Time{
				Value:  params.Extents[0].TimeEnd.Value,
				Format: "RFC3339",
			},
			UssBaseUrl: params.UssBaseUrl,
		},
		Details: scdussv1.OperationalIntentDetails{
			Volumes: &params.Extents,
		},
	}
	intent.Reference.SubscriptionId = d.createImplicitSubscription(params.NewSubscription, id)
	d.Intents[intent.Reference.Id] = intent
	return intent.Reference
}

func (d *InMemoryDSS) createImplicitSubscription(
	implicit *scdussv1.ImplicitSubscriptionParameters,
	dependentIntent scdussv1.EntityID,
) scdussv1.SubscriptionID {
	if implicit == nil {
		return scdussv1.SubscriptionID("")
	}
	subscription := scdussv1.Subscription{
		Id:                          scdussv1.SubscriptionID(uuid.New().String()),
		UssBaseUrl:                  implicit.UssBaseUrl,
		ImplicitSubscription:        new(true),
		NotifyForOperationalIntents: new(true),
		DependentOperationalIntents: &[]scdussv1.EntityID{dependentIntent},
	}
	d.Subscriptions[subscription.Id] = subscription
	return subscription.Id
}

func (d *InMemoryDSS) DeleteOperationalIntentReference(
	ctx context.Context,
	id scdussv1.EntityID,
	ovn scdussv1.EntityOVN,
) (response scdussv1.ChangeOperationalIntentReferenceResponse, err error) {
	mustHaveContext(ctx)
	intent, ok := d.Intents[id]
	if !ok {
		err = fmt.Errorf("dss: operational intent not found %v", id)
		return
	}
	if *intent.Reference.Ovn != ovn {
		err = fmt.Errorf("dss: supplied OVN does not match")
		return
	}
	delete(d.Intents, id)
	return scdussv1.ChangeOperationalIntentReferenceResponse{}, nil
}

func mustHaveContext(ctx context.Context) {
	if ctx == nil {
		panic("inmemory dss: missing Context")
	}
}
