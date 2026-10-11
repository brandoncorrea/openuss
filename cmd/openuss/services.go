package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"bwawan.com/openuss/sdk/api"
	"bwawan.com/openuss/sdk/auth"
	"bwawan.com/openuss/sdk/dss"
	"bwawan.com/openuss/sdk/peer"
	"bwawan.com/openuss/sdk/qualifier/flightplanning"
	"bwawan.com/openuss/sdk/scd"
	"bwawan.com/openuss/sdk/utmclient"
	"github.com/golang-jwt/jwt/v5"
)

type Services struct {
	Intents    scd.IntentStore
	Planning   *flightplanning.Handler
	Authorizer api.Authorizer
}

func NewServices(env Environment) (Services, error) {
	tokens, err := NewTokenSource(env)
	if err != nil {
		return Services{}, err
	}
	intents := scd.NewInMemoryIntentStore()
	planning, err := NewPlanningHandler(env, tokens, intents)
	if err != nil {
		return Services{}, err
	}
	authorizer, err := NewAuthorizer(env)
	if err != nil {
		return Services{}, err
	}
	return Services{
		Intents:    intents,
		Planning:   planning,
		Authorizer: authorizer,
	}, nil
}

func NewAuthorizer(env Environment) (api.Authorizer, error) {
	if env.UnsafeAuthorizer {
		return auth.NewUnsafeAuthorizer(), nil
	}
	if env.OAuthPublicKeyFile == "" {
		return nil, errors.New("OAUTH_PUBLIC_KEY_FILE is required")
	}
	pem, err := os.ReadFile(env.OAuthPublicKeyFile)
	if err != nil {
		return nil, fmt.Errorf("reading OAUTH_PUBLIC_KEY_FILE: %w", err)
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM(pem)
	if err != nil {
		return nil, fmt.Errorf("parsing OAUTH_PUBLIC_KEY_FILE: %w", err)
	}
	return auth.NewJWTAuthorizer(key), nil
}

func NewTokenSource(env Environment) (auth.TokenSource, error) {
	if env.InMemoryTokens {
		return auth.NewInMemoryTokenSource(), nil
	}
	return auth.NewDummyOAuth(env.OAuthEndpoint, env.OAuthSubject, nil)
}

func NewPlanningHandler(env Environment, tokens auth.TokenSource, intents scd.IntentStore) (*flightplanning.Handler, error) {
	if strings.TrimSpace(env.USSBaseURL) == "" {
		return nil, errors.New("USS_BASE_URL is required")
	}

	client := utmclient.New(tokens, nil)
	dssClient := newDSSClient(env, client)
	peerClient := peer.New(client)
	flights := flightplanning.NewInMemoryFlightStore()
	service := scd.New(dssClient, peerClient, intents, env.USSBaseURL)
	return flightplanning.New(service, flights), nil
}

func newDSSClient(env Environment, client *utmclient.Client) dss.Client {
	if env.InMemoryDSS {
		return dss.NewInMemoryDSS()
	}
	return dss.New(env.DSSBaseURL, client)
}
