package scd_test

import (
	"testing"

	"bwawan.com/openuss/sdk/scd"
	"bwawan.com/openuss/sdk/scdtest"
	"github.com/stretchr/testify/require"
)

func TestDeleteOperationalIntentRemovesItFromDSSAndStore(t *testing.T) {
	service, dssClient := newService()
	intent := newCoordinatedIntent(t, service)

	require.NoError(t, service.DeleteOperationalIntent(t.Context(), intent.EntityID))

	_, registered := dssClient.OperationalIntent(intent.EntityID)
	require.False(t, registered)
	_, err := service.Intents.Get(t.Context(), intent.EntityID)
	require.ErrorIs(t, err, scd.ErrNotFound)
}

func TestDeleteOperationalIntentUnknownIDSucceeds(t *testing.T) {
	service, _ := newService()

	err := service.DeleteOperationalIntent(t.Context(), scdtest.NewEntityID())

	require.NoError(t, err)
}
