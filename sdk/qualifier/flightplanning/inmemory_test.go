package flightplanning_test

import (
	"testing"
	"uuid"

	"bwawan.com/openuss/sdk/qualifier/flightplanning"
	"bwawan.com/openuss/sdk/scdtest"
	"github.com/stretchr/testify/require"
)

func TestFlightStoreGetUnknownFlightIsNil(t *testing.T) {
	store := flightplanning.NewInMemoryFlightStore()
	require.Nil(t, store.Get(uuid.New()))
}

func TestFlightStoreDeleteUnknownFlightSucceeds(t *testing.T) {
	store := flightplanning.NewInMemoryFlightStore()
	require.NotPanics(t, func() {
		store.Delete(uuid.New())
	})
}

func TestFlightStoreGetReturnsUpsertedFlight(t *testing.T) {
	store := flightplanning.NewInMemoryFlightStore()
	flight := flightplanning.FlightPlanRecord{
		ID: uuid.New(),
	}
	require.NoError(t, store.Upsert(flight))
	require.Equal(t, flight, *(store.Get(flight.ID)))
}

func TestFlightStoreDeleteRemovesFlight(t *testing.T) {
	store := flightplanning.NewInMemoryFlightStore()
	flight := flightplanning.FlightPlanRecord{
		ID: uuid.New(),
	}
	require.NoError(t, store.Upsert(flight))
	require.Equal(t, flight, *store.Get(flight.ID))
	store.Delete(flight.ID)
	require.Nil(t, store.Get(flight.ID))
}

func TestFlightStoreListIsEmptyWhenNothingIsSaved(t *testing.T) {
	store := flightplanning.NewInMemoryFlightStore()

	require.Empty(t, store.List())
}

func TestFlightStoreListReturnsEveryFlight(t *testing.T) {
	store := flightplanning.NewInMemoryFlightStore()
	first := flightplanning.FlightPlanRecord{ID: uuid.New(), EntityID: scdtest.NewEntityID()}
	second := flightplanning.FlightPlanRecord{ID: uuid.New(), EntityID: scdtest.NewEntityID()}
	require.NoError(t, store.Upsert(first))
	require.NoError(t, store.Upsert(second))

	flights := store.List()

	require.ElementsMatch(t, []flightplanning.FlightPlanRecord{first, second}, flights)
}
