package main_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	main "bwawan.com/openuss/cmd/openuss"
	"bwawan.com/openuss/sdk/auth"
	"bwawan.com/openuss/sdk/dss"
	"bwawan.com/openuss/sdk/httpclient"
	"bwawan.com/openuss/sdk/peer"
	"bwawan.com/openuss/sdk/qualifier/flightplanning"
	"bwawan.com/openuss/sdk/scd"
)

func TestNewServicesSharesOneIntentStoreWithPlanning(t *testing.T) {
	services, err := main.NewServices(main.Environment{
		InMemoryTokens: true,
		InMemoryDSS:    true,
		USSBaseURL:     "the-uss-base-url",
	})

	require.NoError(t, err)
	service := services.Planning.SCD.(*scd.Service)
	require.Same(t, services.Intents, service.Intents)
	require.IsType(t, &auth.UnsafeAuthorizer{}, services.Authorizer)
}

func TestNewServicesHandsTheConfiguredTokenSourceToPlanning(t *testing.T) {
	services, err := main.NewServices(main.Environment{
		OAuthEndpoint: "http://oauth.local",
		OAuthSubject:  "the-oauth-subject",
		DSSBaseURL:    "http://dss.localutm",
		USSBaseURL:    "the-uss-base-url",
	})

	require.NoError(t, err)
	dssClient := services.Planning.SCD.(*scd.Service).DSS.(*dss.UTMClient)
	require.IsType(t, &auth.DummyOAuth{}, dssClient.Client.TokenSource)
	tokens := dssClient.Client.TokenSource.(*auth.DummyOAuth)
	require.Equal(t, "http://oauth.local", tokens.Endpoint.String())
}

func TestNewServicesFailsWhenAServiceCannotBeBuilt(t *testing.T) {
	cases := []struct {
		name string
		env  main.Environment
	}{
		{name: "a token source without an OAuth endpoint", env: main.Environment{USSBaseURL: "the-uss-base-url"}},
		{name: "a planning handler without a USS base URL", env: main.Environment{InMemoryTokens: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := main.NewServices(tc.env)
			require.Error(t, err)
		})
	}
}

func TestNewPlanningHandlerMissingUSSBaseURL(t *testing.T) {
	dummy, _ := auth.NewDummyOAuth("", "", nil)
	intents := scd.NewInMemoryIntentStore()
	env := main.Environment{USSBaseURL: "\r\n\t "}
	_, err := main.NewPlanningHandler(env, dummy, intents)
	require.ErrorContains(t, err, "USS_BASE_URL is required")
}

func TestNewPlanningHandlerWithRealDSS(t *testing.T) {
	dummy, _ := auth.NewDummyOAuth("", "", nil)
	intents := scd.NewInMemoryIntentStore()
	env := main.Environment{
		DSSBaseURL: "http://dss.localutm:8080/blah",
		USSBaseURL: "the-uss-base-url",
	}
	handler, err := main.NewPlanningHandler(env, dummy, intents)
	require.NoError(t, err)
	service := handler.SCD.(*scd.Service)
	dssClient := service.DSS.(*dss.UTMClient)
	require.Equal(t, "http://dss.localutm:8080/blah", dssClient.BaseURL)
	require.Equal(t, dummy, dssClient.Client.TokenSource)
	require.Equal(t, httpclient.DefaultTimeout, dssClient.Client.HTTP.HTTP.Timeout)
	peerClient := service.Peer.(*peer.UTMClient)
	require.Same(t, dssClient.Client, peerClient.Client)
	require.IsType(t, &flightplanning.InMemoryFlightStore{}, handler.Flights)
	require.Same(t, intents, service.Intents)
	require.EqualValues(t, "the-uss-base-url", service.USSBaseURL)
}

func TestNewPlanningHandlerWithMemoryDSS(t *testing.T) {
	dummy, _ := auth.NewDummyOAuth("", "", nil)
	intents := scd.NewInMemoryIntentStore()
	env := main.Environment{
		InMemoryDSS: true,
		USSBaseURL:  "the-uss-base-url",
	}
	handler, err := main.NewPlanningHandler(env, dummy, intents)
	require.NoError(t, err)
	service := handler.SCD.(*scd.Service)
	require.IsType(t, &dss.InMemoryDSS{}, service.DSS)
	require.IsType(t, &peer.UTMClient{}, service.Peer)
	require.IsType(t, &flightplanning.InMemoryFlightStore{}, handler.Flights)
	require.Same(t, intents, service.Intents)
	require.EqualValues(t, "the-uss-base-url", service.USSBaseURL)
}

func TestNewDummyTokenSource(t *testing.T) {
	source, err := main.NewTokenSource(main.Environment{
		OAuthEndpoint: "http://oauth.local",
		OAuthSubject:  "the-oauth-subject",
	})
	require.NoError(t, err)
	dummy := source.(*auth.DummyOAuth)
	require.Equal(t, "http://oauth.local", dummy.Endpoint.String())
	require.Equal(t, "the-oauth-subject", dummy.Subject)
}

func TestNewInMemoryTokenSource(t *testing.T) {
	source, err := main.NewTokenSource(main.Environment{InMemoryTokens: true})
	require.NoError(t, err)
	require.IsType(t, &auth.InMemoryTokenSource{}, source)
}
