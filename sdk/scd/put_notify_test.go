package scd_test

import (
	"testing"

	"bwawan.com/openuss/sdk/dss"
	"bwawan.com/openuss/sdk/peer/peertest"
	"bwawan.com/openuss/sdk/scd"
	"github.com/stretchr/testify/require"
)

func TestCreateOperationalIntentNotifiesPeerSubscriber(t *testing.T) {
	dssClient := dss.NewInMemoryDSS()
	recorder := &peertest.NotificationRecorder{}
	service := scd.New(dssClient, recorder, scd.NewInMemoryIntentStore(), ussBaseURL)
	_, err := service.CreateOperationalIntent(t.Context(), newIntentParamsAt(distantCorner()))
	require.NoError(t, err)
	putPeerIntent(t, dssClient)
	params := newIntentParams()

	intent, err := service.CreateOperationalIntent(t.Context(), params)

	require.NoError(t, err)
	dssIntent, found := dssClient.OperationalIntent(intent.EntityID)
	require.True(t, found)
	require.Len(t, recorder.Notifications, 1)
	notified := recorder.Notifications[0]
	require.EqualValues(t, peerUSSBaseURL, notified.USSBaseURL)
	require.Equal(t, intent.EntityID, notified.Details.OperationalIntentId)
	require.NotNil(t, notified.Details.OperationalIntent)
	require.Equal(t, dssIntent.Reference.Id, notified.Details.OperationalIntent.Reference.Id)
	require.Equal(t, dssIntent.Reference.Manager, notified.Details.OperationalIntent.Reference.Manager)
	require.Equal(t, dssIntent.Reference.UssAvailability, notified.Details.OperationalIntent.Reference.UssAvailability)
	require.Equal(t, dssIntent.Reference.State, notified.Details.OperationalIntent.Reference.State)
	require.Equal(t, dssIntent.Reference.TimeStart, notified.Details.OperationalIntent.Reference.TimeStart)
	require.Equal(t, dssIntent.Reference.TimeEnd, notified.Details.OperationalIntent.Reference.TimeEnd)
	require.Zero(t, notified.Details.OperationalIntent.Reference.Version)
	require.Zero(t, notified.Details.OperationalIntent.Reference.Ovn)
	require.Zero(t, notified.Details.OperationalIntent.Reference.UssBaseUrl)
	require.Zero(t, notified.Details.OperationalIntent.Reference.SubscriptionId)
	require.Zero(t, notified.Details.OperationalIntent.Details)
	require.Zero(t, notified.Details.Subscriptions)
}

func TestCreateOperationalIntentDoesNotNotifyItsOwnSubscription(t *testing.T) {
	recorder := &peertest.NotificationRecorder{}
	service := scd.New(dss.NewInMemoryDSS(), recorder, scd.NewInMemoryIntentStore(), ussBaseURL)

	_, err := service.CreateOperationalIntent(t.Context(), newIntentParams())

	require.NoError(t, err)
	require.Empty(t, recorder.Notifications)
}
