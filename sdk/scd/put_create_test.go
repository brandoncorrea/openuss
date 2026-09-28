package scd_test

import (
	"testing"
	"time"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/scd"
	"github.com/stretchr/testify/require"
)

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
	require.EqualValues(t, ussBaseURL, subscription.UssBaseUrl)
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

	timeStart, err := time.Parse(time.RFC3339Nano, reference.TimeStart.Value)
	require.NoError(t, err)

	timeEnd, err := time.Parse(time.RFC3339Nano, reference.TimeEnd.Value)
	require.NoError(t, err)

	require.Equal(t, timeStart, intent.TimeStart)
	require.Equal(t, timeEnd, intent.TimeEnd)

	stored, err := service.Intents.Get(t.Context(), intent.EntityID)
	require.NoError(t, err)
	require.Equal(t, intent, stored)
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
