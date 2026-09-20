package db_test

import (
	"slices"
	"testing"
	"uuid"

	"bwawan.com/openuss/internal/db"
	"bwawan.com/openuss/sdk/api/scdussv1"
	"github.com/stretchr/testify/require"
)

func TestGetFlightOnEmptyDB(t *testing.T) {
	store := db.NewInMemoryDB()
	require.Nil(t, store.GetFlight(uuid.New()))
}

func TestDeleteFlightOnEmptyDB(t *testing.T) {
	store := db.NewInMemoryDB()
	require.NotPanics(t, func() {
		store.DeleteFlight(uuid.New())
	})
}

func TestGetSavedFlight(t *testing.T) {
	store := db.NewInMemoryDB()
	flight := db.FlightPlan{
		ID: uuid.New(),
	}
	require.NoError(t, store.SaveFlight(flight))
	require.Equal(t, flight, *(store.GetFlight(flight.ID)))
}

func TestDeleteSavedFlight(t *testing.T) {
	store := db.NewInMemoryDB()
	flight := db.FlightPlan{
		ID: uuid.New(),
	}
	require.NoError(t, store.SaveFlight(flight))
	require.Equal(t, flight, *store.GetFlight(flight.ID))
	store.DeleteFlight(flight.ID)
	require.Nil(t, store.GetFlight(flight.ID))
}

func TestGetAllFlights(t *testing.T) {
	store := db.NewInMemoryDB()
	require.Empty(t, slices.Collect(store.GetAllFlights()))

	flight1 := db.FlightPlan{
		ID:       uuid.New(),
		EntityID: scdussv1.EntityID(uuid.New().String()),
	}
	store.SaveFlight(flight1)
	saved := slices.Collect(store.GetAllFlights())
	require.Equal(t, []db.FlightPlan{flight1}, saved)

	flight2 := db.FlightPlan{
		ID:       uuid.New(),
		EntityID: scdussv1.EntityID(uuid.New().String()),
	}
	store.SaveFlight(flight2)
	saved = slices.Collect(store.GetAllFlights())
	require.ElementsMatch(t, []db.FlightPlan{flight1, flight2}, saved)
}
