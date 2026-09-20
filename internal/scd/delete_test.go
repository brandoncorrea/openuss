package scd_test

import (
	"testing"
	"uuid"

	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/dss"
	"bwawan.com/openuss/internal/scd"
	"bwawan.com/openuss/internal/scdtest"
	"github.com/stretchr/testify/require"
)

func newService() (*scd.Service, *dss.InMemoryDSS) {
	dssClient := dss.NewInMemoryDSS()
	return scd.New(dssClient, scd.NewInMemoryIntentStore()), dssClient
}

func newCoordinatedIntent(t *testing.T, service *scd.Service) scd.OperationalIntent {
	t.Helper()
	params := scdussv1.PutOperationalIntentReferenceParameters{
		Extents: scdtest.NewVolumes4D(),
	}
	result, err := service.DSS.PutOperationalIntentReference(t.Context(), newEntityID(), nil, params)
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

	require.NotContains(t, dssClient.Intents, intent.EntityID)
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
	require.Contains(t, dssClient.Intents, intent.EntityID)
	stored, err := service.Intents.Get(t.Context(), intent.EntityID)
	require.NoError(t, err)
	require.Equal(t, intent, stored)
}

func TestDeleteOperationalIntentUnknownIDSucceeds(t *testing.T) {
	service, _ := newService()

	err := service.DeleteOperationalIntent(t.Context(), newEntityID())

	require.NoError(t, err)
}
