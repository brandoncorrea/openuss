package dss

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"uuid"

	"bwawan.com/openuss/sdk/api/scdussv1"
)

type InMemoryDSS struct {
	intents       map[scdussv1.EntityID]scdussv1.OperationalIntent
	subscriptions map[scdussv1.SubscriptionID]scdussv1.Subscription
}

func NewInMemoryDSS() *InMemoryDSS {
	return &InMemoryDSS{
		intents:       map[scdussv1.EntityID]scdussv1.OperationalIntent{},
		subscriptions: map[scdussv1.SubscriptionID]scdussv1.Subscription{},
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
			TimeStart:       *params.Extents[0].TimeStart,
			TimeEnd:         *params.Extents[0].TimeEnd,
			UssBaseUrl:      params.UssBaseUrl,
		},
		Details: scdussv1.OperationalIntentDetails{
			Volumes: &params.Extents,
		},
	}
	intent.Reference.SubscriptionId = d.createImplicitSubscription(params.NewSubscription, id)
	d.intents[intent.Reference.Id] = intent
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
	d.subscriptions[subscription.Id] = subscription
	return subscription.Id
}

func (d *InMemoryDSS) DeleteOperationalIntentReference(
	ctx context.Context,
	id scdussv1.EntityID,
	ovn scdussv1.EntityOVN,
) (response scdussv1.ChangeOperationalIntentReferenceResponse, err error) {
	mustHaveContext(ctx)
	intent, ok := d.intents[id]
	if !ok {
		err = fmt.Errorf("dss: operational intent not found %v", id)
		return
	}
	if *intent.Reference.Ovn != ovn {
		err = fmt.Errorf("dss: supplied OVN does not match")
		return
	}
	delete(d.intents, id)
	return scdussv1.ChangeOperationalIntentReferenceResponse{
		OperationalIntentReference: intent.Reference,
	}, nil
}

func mustHaveContext(ctx context.Context) {
	if ctx == nil {
		panic("inmemory dss: missing Context")
	}
}

func (d *InMemoryDSS) OperationalIntent(id scdussv1.EntityID) (scdussv1.OperationalIntent, bool) {
	intent, found := d.intents[id]
	return intent, found
}

func (d *InMemoryDSS) OperationalIntents() []scdussv1.OperationalIntent {
	return slices.Collect(maps.Values(d.intents))
}

func (d *InMemoryDSS) Subscription(id scdussv1.SubscriptionID) (scdussv1.Subscription, bool) {
	subscription, found := d.subscriptions[id]
	return subscription, found
}
