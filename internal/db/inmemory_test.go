package db_test

import (
	"testing"
	"uuid"

	"bwawan.com/openuss/internal/db"
	"bwawan.com/openuss/sdk/scdtest"
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
	require.Empty(t, store.GetAllFlights())

	flight1 := db.FlightPlan{
		ID:       uuid.New(),
		EntityID: scdtest.NewEntityID(),
	}
	store.SaveFlight(flight1)
	require.Equal(t, []db.FlightPlan{flight1}, store.GetAllFlights())

	flight2 := db.FlightPlan{
		ID:       uuid.New(),
		EntityID: scdtest.NewEntityID(),
	}
	store.SaveFlight(flight2)
	require.ElementsMatch(t, []db.FlightPlan{flight1, flight2}, store.GetAllFlights())
}
