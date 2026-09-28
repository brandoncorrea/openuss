package scd_test

import (
	"testing"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/scd"
	"github.com/stretchr/testify/require"
)

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
