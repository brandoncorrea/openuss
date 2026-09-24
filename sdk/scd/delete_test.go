package scd_test

import (
	"testing"
	"uuid"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/scd"
	"bwawan.com/openuss/sdk/scdtest"
	"github.com/stretchr/testify/require"
)

func newCoordinatedIntent(t *testing.T, service *scd.Service) scd.OperationalIntent {
	t.Helper()
	params := scdussv1.PutOperationalIntentReferenceParameters{
		Extents: scdtest.NewVolumes4D(),
	}
	result, err := service.DSS.PutOperationalIntentReference(t.Context(), scdtest.NewEntityID(), nil, params)
	require.NoError(t, err)

	intent := scd.OperationalIntent{
		EntityID: result.OperationalIntentReference.Id,
		OVN:      *result.OperationalIntentReference.Ovn,
	}
	require.NoError(t, service.Intents.Upsert(t.Context(), intent))
	return intent
}

func TestDeleteOperationalIntentRemovesItFromDSSAndStore(t *testing.T) {
	service, dssClient := newService()
	intent := newCoordinatedIntent(t, service)

	require.NoError(t, service.DeleteOperationalIntent(t.Context(), intent.EntityID))

	_, registered := dssClient.OperationalIntent(intent.EntityID)
	require.False(t, registered)
	_, err := service.Intents.Get(t.Context(), intent.EntityID)
	require.ErrorIs(t, err, scd.ErrNotFound)
}

func TestDeleteOperationalIntentKeepsItWhenDSSRefuses(t *testing.T) {
	service, dssClient := newService()
	intent := newCoordinatedIntent(t, service)
	intent.OVN = scdussv1.EntityOVN(uuid.New().String())
	require.NoError(t, service.Intents.Upsert(t.Context(), intent))

	err := service.DeleteOperationalIntent(t.Context(), intent.EntityID)

	require.Error(t, err)
	_, registered := dssClient.OperationalIntent(intent.EntityID)
	require.True(t, registered)
	stored, err := service.Intents.Get(t.Context(), intent.EntityID)
	require.NoError(t, err)
	require.Equal(t, intent, stored)
}

func TestDeleteOperationalIntentUnknownIDSucceeds(t *testing.T) {
	service, _ := newService()

	err := service.DeleteOperationalIntent(t.Context(), scdtest.NewEntityID())

	require.NoError(t, err)
}
