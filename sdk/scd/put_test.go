package scd_test

import (
	"net/http/httptest"
	"testing"
	"time"
	"uuid"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/auth"
	"bwawan.com/openuss/sdk/dss"
	"bwawan.com/openuss/sdk/dss/dsstest"
	"bwawan.com/openuss/sdk/peer"
	"bwawan.com/openuss/sdk/scd"
	"bwawan.com/openuss/sdk/scdtest"
	"bwawan.com/openuss/sdk/utmclient"
	"github.com/stretchr/testify/require"
)

const ussBaseURL = "http://openuss.localutm"

func newIntentParams() scd.IntentParams {
	return scd.IntentParams{
		Volumes:  scdtest.NewVolumes4D(),
		State:    scdussv1.OperationalIntentState_Accepted,
		Priority: 2,
	}
}

func newServiceFromPeers(t *testing.T, peers []scdussv1.OperationalIntent) *scd.Service {
	t.Helper()
	client := utmclient.New(
		auth.NewInMemoryTokenSource(),
		httptest.NewTestServer(t, dsstest.NewPeerHandler(peers)).Client(),
	)
	return scd.New(
		dss.New("https://dss.localutm", client),
		peer.New(client),
		scd.NewInMemoryIntentStore(),
		ussBaseURL,
	)
}

func newPeerIntent() scdussv1.OperationalIntent {
	return scdussv1.OperationalIntent{
		Reference: scdussv1.OperationalIntentReference{
			Id:         scdtest.NewEntityID(),
			Ovn:        new(scdussv1.EntityOVN(uuid.New().String())),
			UssBaseUrl: scdussv1.OperationalIntentUssBaseURL("http://uss1.localutm"),
		},
	}
}

func newPriority100PeerIntent() scdussv1.OperationalIntent {
	intent := newPeerIntent()
	intent.Details.Priority = new(scdussv1.Priority(100))
	return intent
}

func requireNothingStored(t *testing.T, service *scd.Service) {
	t.Helper()
	intents, err := service.Intents.List(t.Context())
	require.NoError(t, err)
	require.Empty(t, intents)
}

func TestCreateOperationalIntentRegistersItWithTheDSS(t *testing.T) {
	service, dssClient := newService()
	params := newIntentParams()

	_, err := service.CreateOperationalIntent(t.Context(), params)

	require.NoError(t, err)
	require.Len(t, dssClient.OperationalIntents(), 1)
	dssIntent := dssClient.OperationalIntents()[0]
	require.Equal(t, params.Volumes, *dssIntent.Details.Volumes)
	require.Equal(t, scdussv1.OperationalIntentState_Accepted, dssIntent.Reference.State)
	require.EqualValues(t, ussBaseURL, dssIntent.Reference.UssBaseUrl)
	subscription, found := dssClient.Subscription(dssIntent.Reference.SubscriptionId)
	require.True(t, found)
	require.EqualValues(t, "x", subscription.UssBaseUrl)
}

func TestCreateOperationalIntentSendsTheRequestedState(t *testing.T) {
	service, dssClient := newService()
	params := newIntentParams()
	params.State = scdussv1.OperationalIntentState_Activated

	_, err := service.CreateOperationalIntent(t.Context(), params)

	require.NoError(t, err)
	dssIntent := dssClient.OperationalIntents()[0]
	require.Equal(t, scdussv1.OperationalIntentState_Activated, dssIntent.Reference.State)
}

func TestCreateOperationalIntentStoresTheDSSReference(t *testing.T) {
	service, dssClient := newService()
	params := newIntentParams()

	intent, err := service.CreateOperationalIntent(t.Context(), params)

	require.NoError(t, err)
	reference := dssClient.OperationalIntents()[0].Reference
	require.Equal(t, reference.Id, intent.EntityID)
	require.Equal(t, "InMemoryManager", intent.Manager)
	require.Equal(t, scdussv1.UssAvailabilityState_Normal, intent.USSAvailability)
	require.EqualValues(t, 1, intent.Version)
	require.EqualValues(t, 2, intent.Priority)
	require.Equal(t, scdussv1.OperationalIntentState_Accepted, intent.State)
	require.Equal(t, *reference.Ovn, intent.OVN)
	require.Equal(t, reference.SubscriptionId, intent.SubscriptionID)
	require.Equal(t, params.Volumes, intent.Volumes)

	timeStart, _ := time.Parse(time.RFC3339Nano, reference.TimeStart.Value)
	timeEnd, _ := time.Parse(time.RFC3339Nano, reference.TimeEnd.Value)
	require.Equal(t, timeStart, intent.TimeStart)
	require.Equal(t, timeEnd, intent.TimeEnd)

	stored, err := service.Intents.Get(t.Context(), intent.EntityID)
	require.NoError(t, err)
	require.Equal(t, intent, stored)
}

func TestUpdateOperationalIntentKeepsItsEntityIDAndOVN(t *testing.T) {
	service, dssClient := newService()
	params := newIntentParams()
	created, err := service.CreateOperationalIntent(t.Context(), params)
	require.NoError(t, err)

	params.Volumes[0].Volume.AltitudeLower.Value += 1
	updated, err := service.UpdateOperationalIntent(t.Context(), created.EntityID, params)

	require.NoError(t, err)
	require.Equal(t, created.EntityID, updated.EntityID)
	require.NotZero(t, created.OVN)
	require.Equal(t, created.OVN, updated.OVN)

	require.Len(t, dssClient.OperationalIntents(), 1)
	dssIntent, found := dssClient.OperationalIntent(created.EntityID)
	require.True(t, found)
	require.Equal(t, params.Volumes, *dssIntent.Details.Volumes)

	intents, err := service.Intents.List(t.Context())
	require.NoError(t, err)
	require.Equal(t, []scd.OperationalIntent{updated}, intents)
}

func TestCreateOperationalIntentRejectsStartBeyondPlanningHorizon(t *testing.T) {
	service, dssClient := newService()
	params := newIntentParams()
	tooLate := time.Now().Add(time.Hour * 24 * 30).Add(time.Second)
	params.Volumes[0].TimeStart.Value = tooLate.Format(time.RFC3339Nano)
	params.Volumes[0].TimeEnd.Value = tooLate.Add(time.Hour).Format(time.RFC3339Nano)

	_, err := service.CreateOperationalIntent(t.Context(), params)

	require.ErrorIs(t, err, scd.ErrRejected)
	require.Empty(t, dssClient.OperationalIntents())
	requireNothingStored(t, service)
}

func TestCreateOperationalIntentRejectsAnIntentThatHasEnded(t *testing.T) {
	service, dssClient := newService()
	params := newIntentParams()
	oneSecondAgo := time.Now().Add(-time.Second)
	params.Volumes[0].TimeStart.Value = oneSecondAgo.Add(-time.Second).Format(time.RFC3339)
	params.Volumes[0].TimeEnd.Value = oneSecondAgo.Format(time.RFC3339)

	_, err := service.CreateOperationalIntent(t.Context(), params)

	require.ErrorIs(t, err, scd.ErrRejected)
	require.Empty(t, dssClient.OperationalIntents())
	requireNothingStored(t, service)
}

func TestCreateOperationalIntentRejectsWhenAnotherIntentExists(t *testing.T) {
	service, dssClient := newService()
	other := newIntent()
	require.NoError(t, service.Intents.Upsert(t.Context(), other))

	_, err := service.CreateOperationalIntent(t.Context(), newIntentParams())

	require.ErrorIs(t, err, scd.ErrRejected)
	require.Empty(t, dssClient.OperationalIntents())
	intents, err := service.Intents.List(t.Context())
	require.NoError(t, err)
	require.Equal(t, []scd.OperationalIntent{other}, intents)
}

func TestCreateOperationalIntentRetriesWithPeerOVNsWhenKeyIsMissing(t *testing.T) {
	service := newServiceFromPeers(t, []scdussv1.OperationalIntent{newPeerIntent()})

	intent, err := service.CreateOperationalIntent(t.Context(), newIntentParams())

	require.NoError(t, err)
	intents, err := service.Intents.List(t.Context())
	require.NoError(t, err)
	require.Equal(t, []scd.OperationalIntent{intent}, intents)
}

func TestCreateOperationalIntentConflictsWithPeerAtPriority100(t *testing.T) {
	service := newServiceFromPeers(t, []scdussv1.OperationalIntent{newPriority100PeerIntent()})

	_, err := service.CreateOperationalIntent(t.Context(), newIntentParams())

	require.ErrorIs(t, err, scd.ErrConflict)
	requireNothingStored(t, service)
}

func TestCreateOperationalIntentSucceedsWhenPriority100PeerIsActivated(t *testing.T) {
	activated := newPriority100PeerIntent()
	activated.Reference.State = scdussv1.OperationalIntentState_Activated
	service := newServiceFromPeers(t, []scdussv1.OperationalIntent{activated})

	intent, err := service.CreateOperationalIntent(t.Context(), newIntentParams())

	require.NoError(t, err)
	stored, err := service.Intents.Get(t.Context(), intent.EntityID)
	require.NoError(t, err)
	require.Equal(t, intent, stored)
}

func TestUpdateOperationalIntentConflictsWithPeerAtPriority100(t *testing.T) {
	service := newServiceFromPeers(t, []scdussv1.OperationalIntent{newPriority100PeerIntent()})
	existing := scd.OperationalIntent{
		EntityID: scdtest.NewEntityID(),
		OVN:      scdussv1.EntityOVN(uuid.New().String()),
	}
	require.NoError(t, service.Intents.Upsert(t.Context(), existing))

	_, err := service.UpdateOperationalIntent(t.Context(), existing.EntityID, newIntentParams())

	require.ErrorIs(t, err, scd.ErrConflict)
	intents, err := service.Intents.List(t.Context())
	require.NoError(t, err)
	require.Equal(t, []scd.OperationalIntent{existing}, intents)
}
