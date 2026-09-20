package scd_test

import (
	"testing"
	"uuid"

	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/scd"
	"github.com/stretchr/testify/require"
)

func newEntityID() scdussv1.EntityID {
	return scdussv1.EntityID(uuid.New().String())
}

func newIntent() scd.OperationalIntent {
	return scd.OperationalIntent{
		EntityID: newEntityID(),
	}
}

func TestGetUnknownIntentIsNotFound(t *testing.T) {
	store := scd.NewInMemoryIntentStore()

	stored, err := store.Get(t.Context(), newEntityID())

	require.ErrorIs(t, err, scd.ErrNotFound)
	require.Zero(t, stored)
}

func TestGetReturnsUpsertedIntent(t *testing.T) {
	store := scd.NewInMemoryIntentStore()
	intent := newIntent()
	require.NoError(t, store.Upsert(t.Context(), intent))

	stored, err := store.Get(t.Context(), intent.EntityID)

	require.NoError(t, err)
	require.Equal(t, intent, stored)
}

func TestDeleteRemovesIntent(t *testing.T) {
	store := scd.NewInMemoryIntentStore()
	intent := newIntent()
	require.NoError(t, store.Upsert(t.Context(), intent))

	require.NoError(t, store.Delete(t.Context(), intent.EntityID))

	stored, err := store.Get(t.Context(), intent.EntityID)
	require.ErrorIs(t, err, scd.ErrNotFound)
	require.Zero(t, stored)

	intents, err := store.List(t.Context())
	require.NoError(t, err)
	require.Empty(t, intents)
}

func TestListReturnsEveryIntent(t *testing.T) {
	store := scd.NewInMemoryIntentStore()
	first, second := newIntent(), newIntent()
	require.NoError(t, store.Upsert(t.Context(), first))
	require.NoError(t, store.Upsert(t.Context(), second))

	intents, err := store.List(t.Context())

	require.NoError(t, err)
	require.ElementsMatch(t, []scd.OperationalIntent{first, second}, intents)
}

func TestDeleteUnknownIntentSucceeds(t *testing.T) {
	store := scd.NewInMemoryIntentStore()

	require.NoError(t, store.Delete(t.Context(), newEntityID()))
}
