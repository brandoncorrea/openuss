package scd_test

import (
	"net/http"
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
	"bwawan.com/openuss/sdk/scdtest"
	"bwawan.com/openuss/sdk/utmclient"
	"github.com/stretchr/testify/require"
)

const (
	ussBaseURL        = "http://openuss.localutm"
	peerUSSBaseURL    = "http://uss1.localutm"
	squareSideDegrees = 0.001
)

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

func newIntent() scd.OperationalIntent {
	return scd.OperationalIntent{
		EntityID: scdtest.NewEntityID(),
		Volumes:  newSquareVolumes(sharedCorner()),
		Priority: 2,
	}
}

func newOwnIntentAt(corner scdussv1.LatLngPoint) scd.OperationalIntent {
	intent := newIntent()
	intent.Volumes = newSquareVolumes(corner)
	intent.OVN = scdussv1.EntityOVN(uuid.New().String())
	intent.USSBaseURL = ussBaseURL
	return intent
}

func newActivatedOwnIntentAt(corner scdussv1.LatLngPoint) scd.OperationalIntent {
	intent := newOwnIntentAt(corner)
	intent.State = scdussv1.OperationalIntentState_Activated
	return intent
}

func newNonConflictingKnownIntent() scd.OperationalIntent {
	return newOwnIntentAt(distantCorner())
}

func newPeerIntent() scd.OperationalIntent {
	intent := newIntent()
	intent.Priority = 0
	intent.OVN = scdussv1.EntityOVN(uuid.New().String())
	intent.USSBaseURL = peerUSSBaseURL
	return intent
}

func newPeerIntentAt(corner scdussv1.LatLngPoint) scd.OperationalIntent {
	intent := newPeerIntent()
	intent.Volumes = newSquareVolumes(corner)
	return intent
}

func newHighPriorityPeerIntentAt(corner scdussv1.LatLngPoint) scd.OperationalIntent {
	intent := newPeerIntentAt(corner)
	intent.Priority = 100
	return intent
}

func newService() (*scd.Service, *dss.InMemoryDSS) {
	dssClient := dss.NewInMemoryDSS()
	service := scd.New(dssClient, nil, scd.NewInMemoryIntentStore(), ussBaseURL)
	return service, dssClient
}

func newTestUTMClient(t *testing.T, handler http.Handler) *utmclient.Client {
	t.Helper()
	return utmclient.New(
		auth.NewInMemoryTokenSource(),
		httptest.NewTestServer(t, handler).Client(),
	)
}

func newServiceWithPeerClient(
	t *testing.T,
	peerClient peer.Client,
	registered ...scd.OperationalIntent,
) *scd.Service {
	t.Helper()
	dssClient := newTestUTMClient(t, dsstest.NewEcosystemHandler(toDSSIntents(registered...)))
	return scd.New(
		dss.New("https://dss.localutm", dssClient),
		peerClient,
		scd.NewInMemoryIntentStore(),
		ussBaseURL,
	)
}

func newServiceInEcosystem(t *testing.T, registered ...scd.OperationalIntent) *scd.Service {
	t.Helper()
	peerClient := newTestUTMClient(t, dsstest.NewEcosystemHandler(toDSSIntents(registered...)))
	return newServiceWithPeerClient(t, peer.New(peerClient), registered...)
}

func newServiceWithoutPeerLookups(t *testing.T, registered ...scd.OperationalIntent) *scd.Service {
	t.Helper()
	peerClient := newTestUTMClient(t, wiretest.AssertNotCalledHandler(t))
	return newServiceWithPeerClient(t, peer.New(peerClient), registered...)
}

func toDSSIntents(intents ...scd.OperationalIntent) []scdussv1.OperationalIntent {
	return util.Map(intents, func(intent scd.OperationalIntent) scdussv1.OperationalIntent {
		return scdussv1.OperationalIntent{
			Reference: scdussv1.OperationalIntentReference{
				Id:         intent.EntityID,
				Ovn:        new(intent.OVN),
				UssBaseUrl: intent.USSBaseURL,
				State:      intent.State,
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

func putPeerIntent(t *testing.T, dssClient *dss.InMemoryDSS) {
	t.Helper()
	params := scdussv1.PutOperationalIntentReferenceParameters{
		Extents:    newSquareVolumes(sharedCorner()),
		State:      scdussv1.OperationalIntentState_Accepted,
		UssBaseUrl: peerUSSBaseURL,
		NewSubscription: &scdussv1.ImplicitSubscriptionParameters{
			UssBaseUrl: peerUSSBaseURL,
		},
	}
	_, err := dssClient.PutOperationalIntentReference(t.Context(), scdtest.NewEntityID(), nil, params)
	require.NoError(t, err)
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
