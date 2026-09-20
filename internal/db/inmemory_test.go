package db_test

import (
	"testing"
	"uuid"

	"bwawan.com/openuss/internal/db"
	"bwawan.com/openuss/sdk/scdtest"
	"github.com/stretchr/testify/require"
)

func TestGetFlightOnEmptyDB(t *testing.T) {
	store := db.NewInMemoryFlightStore()
	require.Nil(t, store.Get(uuid.New()))
}

func TestDeleteFlightOnEmptyDB(t *testing.T) {
	store := db.NewInMemoryFlightStore()
	require.NotPanics(t, func() {
		store.Delete(uuid.New())
	})
}

func TestGetSavedFlight(t *testing.T) {
	store := db.NewInMemoryFlightStore()
	flight := db.FlightPlan{
		ID: uuid.New(),
	}
	require.NoError(t, store.Upsert(flight))
	require.Equal(t, flight, *(store.Get(flight.ID)))
}

func TestDeleteSavedFlight(t *testing.T) {
	store := db.NewInMemoryFlightStore()
	flight := db.FlightPlan{
		ID: uuid.New(),
	}
	require.NoError(t, store.Upsert(flight))
	require.Equal(t, flight, *store.Get(flight.ID))
	store.Delete(flight.ID)
	require.Nil(t, store.Get(flight.ID))
}

func TestListFlights(t *testing.T) {
	store := db.NewInMemoryFlightStore()
	require.Empty(t, store.List())

	flight1 := db.FlightPlan{
		ID:       uuid.New(),
		EntityID: scdtest.NewEntityID(),
	}
	store.Upsert(flight1)
	require.Equal(t, []db.FlightPlan{flight1}, store.List())

	flight2 := db.FlightPlan{
		ID:       uuid.New(),
		EntityID: scdtest.NewEntityID(),
	}
	store.Upsert(flight2)
	require.ElementsMatch(t, []db.FlightPlan{flight1, flight2}, store.List())
}
