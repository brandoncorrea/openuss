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
	"bwawan.com/openuss/sdk/internal/util"
	"bwawan.com/openuss/sdk/internal/wiretest"
	"bwawan.com/openuss/sdk/peer"
	"bwawan.com/openuss/sdk/scd"
	"bwawan.com/openuss/sdk/utmclient"
	"github.com/stretchr/testify/require"
)

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

const peerUSSBaseURL = "http://uss1.localutm"

func newServiceFromPeers(t *testing.T, peers ...scd.OperationalIntent) *scd.Service {
	t.Helper()
	client := utmclient.New(
		auth.NewInMemoryTokenSource(),
		httptest.NewTestServer(t, dsstest.NewPeerHandler(toDSSIntents(peers...))).Client(),
	)
	return scd.New(
		dss.New("https://dss.localutm", client),
		peer.New(client),
		scd.NewInMemoryIntentStore(),
		ussBaseURL,
	)
}

func newServiceWithoutPeerLookups(t *testing.T, dssIntents ...scd.OperationalIntent) *scd.Service {
	t.Helper()
	dssClient := utmclient.New(
		auth.NewInMemoryTokenSource(),
		httptest.NewTestServer(t, dsstest.NewPeerHandler(toDSSIntents(dssIntents...))).Client(),
	)
	peerClient := utmclient.New(
		auth.NewInMemoryTokenSource(),
		httptest.NewTestServer(t, wiretest.AssertNotCalledHandler(t)).Client(),
	)
	return scd.New(
		dss.New("https://dss.localutm", dssClient),
		peer.New(peerClient),
		scd.NewInMemoryIntentStore(),
		ussBaseURL,
	)
}

func newNonConflictingKnownIntent() scd.OperationalIntent {
	intent := newIntent()
	intent.Volumes = newSquareVolumes(distantCorner())
	intent.OVN = scdussv1.EntityOVN(uuid.New().String())
	intent.USSBaseURL = ussBaseURL
	return intent
}

func toDSSIntents(intents ...scd.OperationalIntent) []scdussv1.OperationalIntent {
	return util.Map(intents, func(intent scd.OperationalIntent) scdussv1.OperationalIntent {
		return scdussv1.OperationalIntent{
			Reference: scdussv1.OperationalIntentReference{
				Id:         intent.EntityID,
				Ovn:        new(intent.OVN),
				UssBaseUrl: intent.USSBaseURL,
			},
			Details: scdussv1.OperationalIntentDetails{
				Volumes:  new(intent.Volumes),
				Priority: new(intent.Priority),
			},
		}
	})
}

func upsertIntents(t *testing.T, service *scd.Service, intents ...scd.OperationalIntent) {
	t.Helper()
	for _, intent := range intents {
		require.NoError(t, service.Intents.Upsert(t.Context(), intent))
	}
}

func newPeerIntent() scd.OperationalIntent {
	intent := newIntent()
	intent.Priority = 0
	intent.OVN = scdussv1.EntityOVN(uuid.New().String())
	intent.USSBaseURL = peerUSSBaseURL
	return intent
}

func newHighPriorityPeerIntentAt(corner scdussv1.LatLngPoint) scd.OperationalIntent {
	intent := newPeerIntent()
	intent.Volumes = newSquareVolumes(corner)
	intent.Priority = 100
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

func TestUpdateOperationalIntentSendsTheRequestedState(t *testing.T) {
	service, dssClient := newService()
	params := newIntentParams()
	intent, err := service.CreateOperationalIntent(t.Context(), params)
	require.NoError(t, err)
	params.State = scdussv1.OperationalIntentState_Activated

	updated, err := service.UpdateOperationalIntent(t.Context(), intent.EntityID, params)

	require.NoError(t, err)
	dssIntent := dssClient.OperationalIntents()[0]
	require.Equal(t, scdussv1.OperationalIntentState_Activated, dssIntent.Reference.State)
	require.Equal(t, scdussv1.OperationalIntentState_Activated, updated.State)
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

func TestCreateOperationalIntentRejectsWhenKnownIntentConflicts(t *testing.T) {
	service, dssClient := newService()
	other := newIntent()
	upsertIntents(t, service, other)

	params := scd.IntentParams{
		Priority: other.Priority,
		Volumes:  other.Volumes,
	}
	intent, err := service.CreateOperationalIntent(t.Context(), params)

	require.ErrorIs(t, err, scd.ErrRejected)
	require.Zero(t, intent)
	require.Empty(t, dssClient.OperationalIntents())
	requireStoredIntents(t, service, other)
}

func TestUpdateOperationalIntentRejectsWhenAnotherIntentConflicts(t *testing.T) {
	service, dssClient := newService()
	first := newIntent()
	second := newIntent()
	upsertIntents(t, service, first, second)

	intent, err := service.UpdateOperationalIntent(t.Context(), second.EntityID, newIntentParams())

	require.ErrorIs(t, err, scd.ErrRejected)
	require.Equal(t, second, intent)
	require.Empty(t, dssClient.OperationalIntents())
	requireStoredIntents(t, service, first, second)
}

func TestCreateOperationalIntentRetriesWithPeerOVNsWhenKeyIsMissing(t *testing.T) {
	service := newServiceFromPeers(t, newPeerIntent())

	intent, err := service.CreateOperationalIntent(t.Context(), newIntentParams())

	require.NoError(t, err)
	requireStoredIntents(t, service, intent)
}

func TestCreateOperationalIntentConflictsWithHighPriorityPeer(t *testing.T) {
	peer := newHighPriorityPeerIntentAt(sharedCorner())
	service := newServiceFromPeers(t, peer)

	intent, err := service.CreateOperationalIntent(t.Context(), newIntentParamsAt(sharedCorner()))

	require.ErrorIs(t, err, scd.ErrConflict)
	require.Zero(t, intent)
	requireNothingStored(t, service)
}

func TestCreateOperationalIntentSucceedsOverOwnLowerPriorityIntent(t *testing.T) {
	service, _ := newService()

	lower, err := service.CreateOperationalIntent(t.Context(), newIntentParams())
	require.NoError(t, err)

	params := newIntentParams()
	params.Priority = lower.Priority + 1
	higher, err := service.CreateOperationalIntent(t.Context(), params)

	require.NoError(t, err)
	requireStoredIntents(t, service, lower, higher)
}

func TestCreateOperationalIntentSucceedsWhenHighPriorityPeerDoesNotIntersect(t *testing.T) {
	peer := newHighPriorityPeerIntentAt(distantCorner())
	service := newServiceFromPeers(t, peer)

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

func TestUpdateActivatedIntentSucceedsWhenItAlreadyConflictsWithHighPriorityPeer(t *testing.T) {
	peer := newHighPriorityPeerIntentAt(sharedCorner())
	service := newServiceFromPeers(t, peer)
	existing := newActivatedIntentAt(sharedCorner())
	upsertIntents(t, service, existing)
	params := newIntentParamsAt(sharedCorner())

	intent, err := service.UpdateOperationalIntent(t.Context(), existing.EntityID, params)

	require.NoError(t, err)
	require.Equal(t, existing.EntityID, intent.EntityID)
	require.Equal(t, params.Volumes, intent.Volumes)
	requireStoredIntents(t, service, intent)
}

func TestUpdateAcceptedIntentConflictsWhenItAlreadyConflictsWithHighPriorityPeer(t *testing.T) {
	peer := newHighPriorityPeerIntentAt(sharedCorner())
	service := newServiceFromPeers(t, peer)
	existing := newActivatedIntentAt(sharedCorner())
	existing.State = scdussv1.OperationalIntentState_Accepted
	upsertIntents(t, service, existing)

	intent, err := service.UpdateOperationalIntent(t.Context(), existing.EntityID, newIntentParamsAt(sharedCorner()))

	require.ErrorIs(t, err, scd.ErrConflict)
	require.Equal(t, existing, intent)
	requireStoredIntents(t, service, existing)
}

func TestUpdateActivatedIntentConflictsWhenMovingIntoHighPriorityPeer(t *testing.T) {
	peer := newHighPriorityPeerIntentAt(sharedCorner())
	service := newServiceFromPeers(t, peer)
	existing := newActivatedIntentAt(distantCorner())
	upsertIntents(t, service, existing)

	intent, err := service.UpdateOperationalIntent(t.Context(), existing.EntityID, newIntentParamsAt(sharedCorner()))

	require.ErrorIs(t, err, scd.ErrConflict)
	require.Equal(t, existing, intent)
	requireStoredIntents(t, service, existing)
}

func TestCreateOperationalIntentSucceedsWithoutPeerLookupWhenKeyCoversKnownIntents(t *testing.T) {
	intents := []scd.OperationalIntent{newNonConflictingKnownIntent(), newNonConflictingKnownIntent()}
	service := newServiceWithoutPeerLookups(t, intents...)
	upsertIntents(t, service, intents...)

	intent, err := service.CreateOperationalIntent(t.Context(), newIntentParams())

	require.NoError(t, err)
	requireStoredIntents(t, service, append(intents, intent)...)
}

func TestCreateOperationalIntentRetriesWithKnownAndPeerOVNs(t *testing.T) {
	known := newNonConflictingKnownIntent()
	service := newServiceFromPeers(t, known, newPeerIntent())
	upsertIntents(t, service, known)

	intent, err := service.CreateOperationalIntent(t.Context(), newIntentParams())

	require.NoError(t, err)
	requireStoredIntents(t, service, known, intent)
}

func TestCreateActivatedIntentConflicts(t *testing.T) {
	service, _ := newService()
	params := newIntentParams()
	params.State = scdussv1.OperationalIntentState_Activated
	intent, err := service.CreateOperationalIntent(t.Context(), params)

	require.Zero(t, intent)
	require.ErrorIs(t, err, scd.ErrConflict)
	requireNothingStored(t, service)
}

func TestCreateNonconformingIntentIsNotSupported(t *testing.T) {
	service, _ := newService()
	params := newIntentParams()
	params.State = scdussv1.OperationalIntentState_Nonconforming
	intent, err := service.CreateOperationalIntent(t.Context(), params)

	require.Zero(t, intent)
	require.ErrorIs(t, err, scd.ErrNotSupported)
	requireNothingStored(t, service)
}
