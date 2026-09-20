package dss_test

import (
	"testing"
	"uuid"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/dss"
	"bwawan.com/openuss/sdk/scdtest"
	"github.com/stretchr/testify/require"
)

func putIntent(t *testing.T, authority *dss.InMemoryDSS) scdussv1.OperationalIntentReference {
	t.Helper()
	id := scdtest.NewEntityID()
	params := scdussv1.PutOperationalIntentReferenceParameters{
		Extents: scdtest.NewVolumes4D(),
		NewSubscription: &scdussv1.ImplicitSubscriptionParameters{
			UssBaseUrl: "http://uss.example.com",
		},
	}
	result, err := authority.PutOperationalIntentReference(t.Context(), id, nil, params)
	require.NoError(t, err)
	return result.OperationalIntentReference
}

func TestInMemoryDSSCreatesAnOperationalIntent(t *testing.T) {
	authority := dss.NewInMemoryDSS()
	id := scdtest.NewEntityID()
	params := scdussv1.PutOperationalIntentReferenceParameters{
		State:      scdussv1.OperationalIntentState_Accepted,
		UssBaseUrl: "the-uss-base-url",
		NewSubscription: &scdussv1.ImplicitSubscriptionParameters{
			UssBaseUrl: "the-sub-base-url",
		},
		Extents: []scdussv1.Volume4D{
			{
				TimeStart: &scdussv1.Time{
					Value:  "the-start-value",
					Format: "the-start-format",
				},
				TimeEnd: &scdussv1.Time{
					Value:  "the-end-value",
					Format: "the-end-format",
				},
			},
		},
	}

	intent, err := authority.PutOperationalIntentReference(t.Context(), id, nil, params)
	require.NoError(t, err)

	reference := intent.OperationalIntentReference
	require.Equal(t, id, reference.Id)
	require.Equal(t, "InMemoryManager", reference.Manager)
	require.Equal(t, scdussv1.UssAvailabilityState_Normal, reference.UssAvailability)
	require.EqualValues(t, 1, reference.Version)
	require.Equal(t, scdussv1.OperationalIntentState_Accepted, reference.State)
	require.NotZero(t, reference.Ovn)
	require.Equal(t, "the-start-value", reference.TimeStart.Value)
	require.Equal(t, "the-start-format", reference.TimeStart.Format)
	require.Equal(t, "the-end-value", reference.TimeEnd.Value)
	require.Equal(t, "the-end-format", reference.TimeEnd.Format)
	require.EqualValues(t, "the-uss-base-url", reference.UssBaseUrl)
	require.NotZero(t, reference.SubscriptionId)

	saved, found := authority.OperationalIntent(reference.Id)
	require.True(t, found)
	require.Equal(t, params.Extents, *saved.Details.Volumes)

	sub, found := authority.Subscription(reference.SubscriptionId)
	require.NotZero(t, sub)
	require.EqualValues(t, "the-sub-base-url", sub.UssBaseUrl)
	require.True(t, *sub.ImplicitSubscription)
	require.True(t, *sub.NotifyForOperationalIntents)
	require.Equal(t, []scdussv1.EntityID{reference.Id}, *sub.DependentOperationalIntents)
}

func TestInMemoryDSSCreatePanicsWithoutContext(t *testing.T) {
	authority := dss.NewInMemoryDSS()
	id := scdtest.NewEntityID()
	params := scdussv1.PutOperationalIntentReferenceParameters{
		Extents: scdtest.NewVolumes4D(),
	}

	require.Panics(t, func() {
		authority.PutOperationalIntentReference(nil, id, nil, params)
	})
}

func TestInMemoryDSSUpdatesAnOperationalIntent(t *testing.T) {
	authority := dss.NewInMemoryDSS()
	reference := putIntent(t, authority)
	saved, _ := authority.OperationalIntent(reference.Id)
	params := scdussv1.PutOperationalIntentReferenceParameters{
		Extents: *saved.Details.Volumes,
	}

	intent, err := authority.PutOperationalIntentReference(t.Context(), reference.Id, reference.Ovn, params)
	require.NoError(t, err)
	require.Equal(t, *intent.OperationalIntentReference.Ovn, *saved.Reference.Ovn)
}

func TestInMemoryDSSCreatesAnOperationalIntentWithoutImplicitSubscription(t *testing.T) {
	authority := dss.NewInMemoryDSS()
	id := scdtest.NewEntityID()
	params := scdussv1.PutOperationalIntentReferenceParameters{
		Extents: scdtest.NewVolumes4D(),
	}

	intent, err := authority.PutOperationalIntentReference(t.Context(), id, nil, params)
	require.NoError(t, err)

	saved, found := authority.OperationalIntent(intent.OperationalIntentReference.Id)
	require.True(t, found)
	require.Zero(t, saved.Reference.SubscriptionId)

	sub, found := authority.Subscription(saved.Reference.SubscriptionId)
	require.False(t, found)
	require.Zero(t, sub)
}

func TestInMemoryDSSFindsAPutOperationalIntent(t *testing.T) {
	authority := dss.NewInMemoryDSS()
	reference := putIntent(t, authority)

	intent, found := authority.OperationalIntent(reference.Id)

	require.True(t, found)
	require.Equal(t, reference, intent.Reference)
}

func TestInMemoryDSSDoesNotFindAnUnknownOperationalIntent(t *testing.T) {
	authority := dss.NewInMemoryDSS()
	putIntent(t, authority)

	intent, found := authority.OperationalIntent(scdtest.NewEntityID())

	require.False(t, found)
	require.Zero(t, intent)
}

func TestInMemoryDSSListsEveryOperationalIntent(t *testing.T) {
	authority := dss.NewInMemoryDSS()
	first, second := putIntent(t, authority), putIntent(t, authority)

	intents := authority.OperationalIntents()

	require.Len(t, intents, 2)
	require.ElementsMatch(t,
		[]scdussv1.EntityID{first.Id, second.Id},
		[]scdussv1.EntityID{intents[0].Reference.Id, intents[1].Reference.Id})
}

func TestInMemoryDSSFindsTheImplicitSubscription(t *testing.T) {
	authority := dss.NewInMemoryDSS()
	reference := putIntent(t, authority)

	subscription, found := authority.Subscription(reference.SubscriptionId)

	require.True(t, found)
	require.EqualValues(t, "http://uss.example.com", subscription.UssBaseUrl)
}

func TestInMemoryDSSDeletesOperationalIntent(t *testing.T) {
	authority := dss.NewInMemoryDSS()
	reference := putIntent(t, authority)
	saved, _ := authority.OperationalIntent(reference.Id)

	result, err := authority.DeleteOperationalIntentReference(t.Context(), reference.Id, *reference.Ovn)
	require.NoError(t, err)
	require.Equal(t, saved.Reference, result.OperationalIntentReference)

	saved, found := authority.OperationalIntent(reference.Id)
	require.False(t, found)
	require.Zero(t, saved)
}

func TestInMemoryDSSCannotDeleteOperationalIntentWithMismatchedOVN(t *testing.T) {
	authority := dss.NewInMemoryDSS()
	reference := putIntent(t, authority)
	badOVN := scdussv1.EntityOVN(uuid.New().String())

	result, err := authority.DeleteOperationalIntentReference(t.Context(), reference.Id, badOVN)
	require.EqualError(t, err, "dss: supplied OVN does not match")
	require.Zero(t, result.OperationalIntentReference)

	saved, found := authority.OperationalIntent(reference.Id)
	require.True(t, found)
	require.Equal(t, reference, saved.Reference)
}

func TestInMemoryDSSCannotDeleteOperationalIntentNotFound(t *testing.T) {
	authority := dss.NewInMemoryDSS()
	entityID := scdtest.NewEntityID()
	ovn := scdussv1.EntityOVN(uuid.New().String())

	result, err := authority.DeleteOperationalIntentReference(t.Context(), entityID, ovn)
	require.EqualError(t, err, "dss: operational intent not found "+string(entityID))
	require.Zero(t, result.OperationalIntentReference)
}

func TestInMemoryDSSDeletePanicsWithoutContext(t *testing.T) {
	authority := dss.NewInMemoryDSS()
	reference := putIntent(t, authority)

	require.Panics(t, func() {
		authority.DeleteOperationalIntentReference(nil, reference.Id, *reference.Ovn)
	})
}
