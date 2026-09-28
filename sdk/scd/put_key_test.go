package scd_test

import (
	"testing"

	"bwawan.com/openuss/sdk/scd"
	"github.com/stretchr/testify/require"
)

func TestCreateOperationalIntentRetriesWithPeerOVNsWhenKeyIsMissing(t *testing.T) {
	service := newServiceInEcosystem(t, newPeerIntent())

	intent, err := service.CreateOperationalIntent(t.Context(), newIntentParams())

	require.NoError(t, err)
	requireStoredIntents(t, service, intent)
}

func TestCreateOperationalIntentSucceedsWithoutPeerLookupWhenKeyCoversKnownIntents(t *testing.T) {
	intents := []scd.OperationalIntent{newNonConflictingKnownIntent(), newNonConflictingKnownIntent()}
	service := newServiceWithoutPeerLookups(t, intents...)
	upsertIntents(t, service, intents...)

	intent, err := service.CreateOperationalIntent(t.Context(), newIntentParams())

	require.NoError(t, err)
	requireStoredIntents(t, service, append(intents, intent)...)
}

func TestCreateOperationalIntentRetriesWithKnownAndPeerOVNs(t *testing.T) {
	known := newNonConflictingKnownIntent()
	service := newServiceInEcosystem(t, known, newPeerIntent())
	upsertIntents(t, service, known)

	intent, err := service.CreateOperationalIntent(t.Context(), newIntentParams())

	require.NoError(t, err)
	requireStoredIntents(t, service, known, intent)
}
