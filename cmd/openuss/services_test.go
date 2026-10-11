package main_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
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
		InMemoryDSS:      true,
		InMemoryTokens:   true,
		UnsafeAuthorizer: true,
		USSBaseURL:       "the-uss-base-url",
	})

	require.NoError(t, err)
	service := services.Planning.SCD.(*scd.Service)
	require.Same(t, services.Intents, service.Intents)
}

func TestNewServicesHandsTheConfiguredTokenSourceToPlanning(t *testing.T) {
	services, err := main.NewServices(main.Environment{
		DSSBaseURL:       "http://dss.localutm",
		OAuthEndpoint:    "http://oauth.local",
		OAuthSubject:     "the-oauth-subject",
		UnsafeAuthorizer: true,
		USSBaseURL:       "the-uss-base-url",
	})

	require.NoError(t, err)
	dssClient := services.Planning.SCD.(*scd.Service).DSS.(*dss.UTMClient)
	require.IsType(t, &auth.DummyOAuth{}, dssClient.Client.TokenSource)
	tokens := dssClient.Client.TokenSource.(*auth.DummyOAuth)
	require.Equal(t, "http://oauth.local", tokens.Endpoint.String())
}

func TestNewServicesUsesTheConfiguredAuthorizer(t *testing.T) {
	path, _ := writePublicKey(t)

	services, err := main.NewServices(main.Environment{
		InMemoryTokens:     true,
		OAuthPublicKeyFile: path,
		USSBaseURL:         "the-uss-base-url",
	})

	require.NoError(t, err)
	require.IsType(t, &auth.JWTAuthorizer{}, services.Authorizer)
}

func TestNewServicesFailsWhenAServiceCannotBeBuilt(t *testing.T) {
	cases := []struct {
		name string
		env  main.Environment
	}{
		{
			name: "a token source without an OAuth endpoint",
			env:  main.Environment{USSBaseURL: "the-uss-base-url", UnsafeAuthorizer: true},
		},
		{
			name: "a planning handler without a USS base URL",
			env:  main.Environment{InMemoryTokens: true, UnsafeAuthorizer: true},
		},
		{
			name: "an authorizer without a usable key",
			env:  main.Environment{InMemoryTokens: true, USSBaseURL: "the-uss-base-url"},
		},
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

func TestNewUnsafeAuthorizer(t *testing.T) {
	authorizer, err := main.NewAuthorizer(main.Environment{UnsafeAuthorizer: true})
	require.NoError(t, err)
	require.IsType(t, &auth.UnsafeAuthorizer{}, authorizer)
}

func writePublicKey(t *testing.T) (string, *rsa.PublicKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	encoded, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "auth.pem")
	block := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: encoded})
	require.NoError(t, os.WriteFile(path, block, 0o600))
	return path, &key.PublicKey
}

func TestNewJWTAuthorizer(t *testing.T) {
	path, key := writePublicKey(t)

	authorizer, err := main.NewAuthorizer(main.Environment{OAuthPublicKeyFile: path})

	require.NoError(t, err)
	require.IsType(t, &auth.JWTAuthorizer{}, authorizer)
	require.Equal(t, key, authorizer.(*auth.JWTAuthorizer).PublicKey)
}

func TestNewAuthorizerRequiresAUsableKey(t *testing.T) {
	notAKey := filepath.Join(t.TempDir(), "not-a-key.pem")
	require.NoError(t, os.WriteFile(notAKey, []byte("not a key"), 0o600))
	cases := []struct {
		name    string
		path    string
		wantErr string
	}{
		{
			name:    "no key file configured",
			path:    "",
			wantErr: "OAUTH_PUBLIC_KEY_FILE is required",
		},
		{
			name:    "a key file that does not exist",
			path:    filepath.Join(t.TempDir(), "missing.pem"),
			wantErr: "reading OAUTH_PUBLIC_KEY_FILE",
		},
		{
			name:    "a key file that is not a PEM public key",
			path:    notAKey,
			wantErr: "parsing OAUTH_PUBLIC_KEY_FILE",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := main.NewAuthorizer(main.Environment{OAuthPublicKeyFile: tc.path})
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}
