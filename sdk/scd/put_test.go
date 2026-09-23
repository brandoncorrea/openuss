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

const blockingPriority = 100

const squareSideDegrees = 0.001

func sharedCorner() scdussv1.LatLngPoint {
	return scdussv1.LatLngPoint{Lng: -80.6, Lat: 37.2}
}

func distantCorner() scdussv1.LatLngPoint {
	return scdussv1.LatLngPoint{Lng: -80.5, Lat: 37.2}
}

func newSquareVolumes(corner scdussv1.LatLngPoint) []scdussv1.Volume4D {
	volumes := scdtest.NewVolumes4D()
	volumes[0].Volume.OutlineCircle = nil
	volumes[0].Volume.OutlinePolygon = &scdussv1.Polygon{
		Vertices: []scdussv1.LatLngPoint{
			corner,
			{Lng: corner.Lng + squareSideDegrees, Lat: corner.Lat},
			{Lng: corner.Lng + squareSideDegrees, Lat: corner.Lat + squareSideDegrees},
			{Lng: corner.Lng, Lat: corner.Lat + squareSideDegrees},
		},
	}
	return volumes
}

func newIntentParams() scd.IntentParams {
	return newIntentParamsAt(sharedCorner())
}

func newIntentParamsAt(corner scdussv1.LatLngPoint) scd.IntentParams {
	return scd.IntentParams{
		Volumes:  newSquareVolumes(corner),
		State:    scdussv1.OperationalIntentState_Accepted,
		Priority: 2,
	}
}

func newEndedIntentParams() scd.IntentParams {
	params := newIntentParams()
	oneSecondAgo := time.Now().Add(-time.Second)
	params.Volumes[0].TimeStart.Value = oneSecondAgo.Add(-time.Second).Format(time.RFC3339)
	params.Volumes[0].TimeEnd.Value = oneSecondAgo.Format(time.RFC3339)
	return params
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
		Details: scdussv1.OperationalIntentDetails{
			Volumes:  new(newSquareVolumes(sharedCorner())),
			Priority: new(scdussv1.Priority(0)),
		},
	}
}

func newPriority100PeerIntentAt(corner scdussv1.LatLngPoint) scdussv1.OperationalIntent {
	intent := newPeerIntent()
	intent.Details.Volumes = new(newSquareVolumes(corner))
	intent.Details.Priority = new(scdussv1.Priority(blockingPriority))
	return intent
}

func newActivatedIntentAt(corner scdussv1.LatLngPoint) scd.OperationalIntent {
	intent := newIntent()
	intent.State = scdussv1.OperationalIntentState_Activated
	intent.Volumes = newSquareVolumes(corner)
	return intent
}

func requireNothingStored(t *testing.T, service *scd.Service) {
	t.Helper()
	intents, err := service.Intents.List(t.Context())
	require.NoError(t, err)
	require.Empty(t, intents)
}

func requireStoredIntents(t *testing.T, service *scd.Service, expected ...scd.OperationalIntent) {
	t.Helper()
	intents, err := service.Intents.List(t.Context())
	require.NoError(t, err)
	require.ElementsMatch(t, expected, intents)
}

func TestCreateOperationalIntentRegistersItWithTheDSS(t *testing.T) {
	service, dssClient := newService()
	params := newIntentParams()

	_, err := service.CreateOperationalIntent(t.Context(), params)

	require.NoError(t, err)
	require.Len(t, dssClient.OperationalIntents(), 1)
	dssIntent := dssClient.OperationalIntents()[0]
	require.Equal(t, params.Volumes, *dssIntent.Details.Volumes)
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
	requireStoredIntents(t, service, updated)
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

	intent, err := service.CreateOperationalIntent(t.Context(), newEndedIntentParams())

	require.ErrorIs(t, err, scd.ErrRejected)
	require.Zero(t, intent)
	require.Empty(t, dssClient.OperationalIntents())
	requireNothingStored(t, service)
}

func TestCreateOperationalIntentRejectsWhenAnotherIntentExists(t *testing.T) {
	service, dssClient := newService()
	other := newIntent()
	require.NoError(t, service.Intents.Upsert(t.Context(), other))

	intent, err := service.CreateOperationalIntent(t.Context(), newIntentParams())

	require.ErrorIs(t, err, scd.ErrRejected)
	require.Zero(t, intent)
	require.Empty(t, dssClient.OperationalIntents())
	requireStoredIntents(t, service, other)
}

func TestUpdateOperationalIntentRejectsWhenAnotherIntentExists(t *testing.T) {
	service, dssClient := newService()
	first := newIntent()
	second := newIntent()
	require.NoError(t, service.Intents.Upsert(t.Context(), first))
	require.NoError(t, service.Intents.Upsert(t.Context(), second))

	intent, err := service.UpdateOperationalIntent(t.Context(), second.EntityID, newIntentParams())

	require.ErrorIs(t, err, scd.ErrRejected)
	require.Equal(t, second, intent)
	require.Empty(t, dssClient.OperationalIntents())
	requireStoredIntents(t, service, first, second)
}

func TestCreateOperationalIntentRetriesWithPeerOVNsWhenKeyIsMissing(t *testing.T) {
	service := newServiceFromPeers(t, []scdussv1.OperationalIntent{newPeerIntent()})

	intent, err := service.CreateOperationalIntent(t.Context(), newIntentParams())

	require.NoError(t, err)
	requireStoredIntents(t, service, intent)
}

func TestCreateOperationalIntentConflictsWithPeerAtPriority100(t *testing.T) {
	peer := newPriority100PeerIntentAt(sharedCorner())
	service := newServiceFromPeers(t, []scdussv1.OperationalIntent{peer})

	intent, err := service.CreateOperationalIntent(t.Context(), newIntentParamsAt(sharedCorner()))

	require.ErrorIs(t, err, scd.ErrConflict)
	require.Zero(t, intent)
	requireNothingStored(t, service)
}

func TestCreateOperationalIntentSucceedsWhenPriority100PeerDoesNotIntersect(t *testing.T) {
	peer := newPriority100PeerIntentAt(distantCorner())
	service := newServiceFromPeers(t, []scdussv1.OperationalIntent{peer})

	intent, err := service.CreateOperationalIntent(t.Context(), newIntentParamsAt(sharedCorner()))

	require.NoError(t, err)
	requireStoredIntents(t, service, intent)
}

func TestUpdateOperationalIntentRejectsAnIntentThatHasEnded(t *testing.T) {
	service, dssClient := newService()
	existing := newCoordinatedIntent(t, service)

	intent, err := service.UpdateOperationalIntent(t.Context(), existing.EntityID, newEndedIntentParams())

	require.ErrorIs(t, err, scd.ErrRejected)
	require.Equal(t, existing, intent)
	dssIntent, found := dssClient.OperationalIntent(existing.EntityID)
	require.True(t, found)
	require.Equal(t, existing.OVN, *dssIntent.Reference.Ovn)
	requireStoredIntents(t, service, existing)
}

func TestUpdateActivatedIntentSucceedsWhenItAlreadyConflictsWithPriority100Peer(t *testing.T) {
	peer := newPriority100PeerIntentAt(sharedCorner())
	service := newServiceFromPeers(t, []scdussv1.OperationalIntent{peer})
	existing := newActivatedIntentAt(sharedCorner())
	require.NoError(t, service.Intents.Upsert(t.Context(), existing))
	params := newIntentParamsAt(sharedCorner())

	intent, err := service.UpdateOperationalIntent(t.Context(), existing.EntityID, params)

	require.NoError(t, err)
	require.Equal(t, existing.EntityID, intent.EntityID)
	require.Equal(t, params.Volumes, intent.Volumes)
	requireStoredIntents(t, service, intent)
}

func TestUpdateAcceptedIntentConflictsWhenItAlreadyConflictsWithPriority100Peer(t *testing.T) {
	peer := newPriority100PeerIntentAt(sharedCorner())
	service := newServiceFromPeers(t, []scdussv1.OperationalIntent{peer})
	existing := newActivatedIntentAt(sharedCorner())
	existing.State = scdussv1.OperationalIntentState_Accepted
	require.NoError(t, service.Intents.Upsert(t.Context(), existing))

	intent, err := service.UpdateOperationalIntent(t.Context(), existing.EntityID, newIntentParamsAt(sharedCorner()))

	require.ErrorIs(t, err, scd.ErrConflict)
	require.Equal(t, existing, intent)
	requireStoredIntents(t, service, existing)
}

func TestUpdateActivatedIntentConflictsWhenMovingIntoPriority100Peer(t *testing.T) {
	peer := newPriority100PeerIntentAt(sharedCorner())
	service := newServiceFromPeers(t, []scdussv1.OperationalIntent{peer})
	existing := newActivatedIntentAt(distantCorner())
	require.NoError(t, service.Intents.Upsert(t.Context(), existing))

	intent, err := service.UpdateOperationalIntent(t.Context(), existing.EntityID, newIntentParamsAt(sharedCorner()))

	require.ErrorIs(t, err, scd.ErrConflict)
	require.Equal(t, existing, intent)
	requireStoredIntents(t, service, existing)
}
