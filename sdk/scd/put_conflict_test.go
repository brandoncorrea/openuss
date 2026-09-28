package scd_test

import (
	"testing"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/peer/peertest"
	"bwawan.com/openuss/sdk/scd"
	"github.com/stretchr/testify/require"
)

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

func TestCreateOperationalIntentConflictsWithHighPriorityPeer(t *testing.T) {
	peer := newHighPriorityPeerIntentAt(sharedCorner())
	service := newServiceInEcosystem(t, peer)

	intent, err := service.CreateOperationalIntent(t.Context(), newIntentParamsAt(sharedCorner()))

	require.ErrorIs(t, err, scd.ErrConflict)
	require.Zero(t, intent)
	requireNothingStored(t, service)
}

func TestCreateOperationalIntentConflictsWithPeerWithInvalidFlyingState(t *testing.T) {
	peer := newPeerIntentAt(distantCorner())
	peer.State = "Flying"
	service := newServiceInEcosystem(t, peer)

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
	service := newServiceInEcosystem(t, peer)

	intent, err := service.CreateOperationalIntent(t.Context(), newIntentParamsAt(sharedCorner()))

	require.NoError(t, err)
	requireStoredIntents(t, service, intent)
}

func TestUpdateActivatedIntentSucceedsWhenItAlreadyConflictsWithHighPriorityPeer(t *testing.T) {
	peer := newHighPriorityPeerIntentAt(sharedCorner())
	service := newServiceInEcosystem(t, peer)
	existing := newActivatedOwnIntentAt(sharedCorner())
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
	service := newServiceInEcosystem(t, peer)
	existing := newOwnIntentAt(sharedCorner())
	existing.State = scdussv1.OperationalIntentState_Accepted
	upsertIntents(t, service, existing)

	intent, err := service.UpdateOperationalIntent(t.Context(), existing.EntityID, newIntentParamsAt(sharedCorner()))

	require.ErrorIs(t, err, scd.ErrConflict)
	require.Equal(t, existing, intent)
	requireStoredIntents(t, service, existing)
}

func TestUpdateActivatedIntentConflictsWhenMovingIntoHighPriorityPeer(t *testing.T) {
	peer := newHighPriorityPeerIntentAt(sharedCorner())
	service := newServiceInEcosystem(t, peer)
	existing := newActivatedOwnIntentAt(distantCorner())
	upsertIntents(t, service, existing)

	intent, err := service.UpdateOperationalIntent(t.Context(), existing.EntityID, newIntentParamsAt(sharedCorner()))

	require.ErrorIs(t, err, scd.ErrConflict)
	require.Equal(t, existing, intent)
	requireStoredIntents(t, service, existing)
}

func TestCreateOperationalIntentConflictsWhenPeerReturnsError(t *testing.T) {
	service := newServiceWithPeerClient(t, peertest.UnreachablePeer{}, newPeerIntent())

	intent, err := service.CreateOperationalIntent(t.Context(), newIntentParams())

	require.ErrorIs(t, err, scd.ErrConflict)
	require.Zero(t, intent)
	requireNothingStored(t, service)
}
