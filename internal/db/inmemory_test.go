package db_test

import (
	"slices"
	"testing"
	"uuid"

	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/db"
	"github.com/stretchr/testify/require"
)

func TestGetIntentOnEmptyDB(t *testing.T) {
	store := db.NewInMemoryDB()
	id := scdussv1.EntityID(uuid.New().String())
	require.Nil(t, store.GetIntent(id))
}

func TestDeleteIntentOnEmptyDB(t *testing.T) {
	store := db.NewInMemoryDB()
	id := scdussv1.EntityID(uuid.New().String())
	require.NotPanics(t, func() {
		store.DeleteIntent(id)
	})
}

func TestGetSavedIntent(t *testing.T) {
	store := db.NewInMemoryDB()
	intent := db.OperationalIntent{
		EntityID: scdussv1.EntityID(uuid.New().String()),
	}
	require.NoError(t, store.SaveIntent(intent))
	require.Equal(t, intent, *(store.GetIntent(intent.EntityID)))
}

func TestDeleteSavedIntent(t *testing.T) {
	store := db.NewInMemoryDB()
	id := scdussv1.EntityID(uuid.New().String())
	intent := db.OperationalIntent{
		EntityID: id,
	}
	require.NoError(t, store.SaveIntent(intent))
	require.Equal(t, intent, *store.GetIntent(id))
	store.DeleteIntent(id)
	require.Nil(t, store.GetIntent(id))
}

func TestGetAllIntents(t *testing.T) {
	store := db.NewInMemoryDB()
	require.Empty(t, slices.Collect(store.GetAllIntents()))

	intent1 := db.OperationalIntent{
		EntityID: scdussv1.EntityID(uuid.New().String()),
	}
	store.SaveIntent(intent1)
	saved := slices.Collect(store.GetAllIntents())
	require.Equal(t, []db.OperationalIntent{intent1}, saved)

	intent2 := db.OperationalIntent{
		EntityID: scdussv1.EntityID(uuid.New().String()),
	}
	store.SaveIntent(intent2)
	saved = slices.Collect(store.GetAllIntents())
	require.ElementsMatch(t, []db.OperationalIntent{intent1, intent2}, saved)
}

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
		Id: uuid.New(),
	}
	require.NoError(t, store.SaveFlight(flight))
	require.Equal(t, flight, *(store.GetFlight(flight.Id)))
}

func TestDeleteSavedFlight(t *testing.T) {
	store := db.NewInMemoryDB()
	flight := db.FlightPlan{
		Id: uuid.New(),
	}
	require.NoError(t, store.SaveFlight(flight))
	require.Equal(t, flight, *store.GetFlight(flight.Id))
	store.DeleteFlight(flight.Id)
	require.Nil(t, store.GetFlight(flight.Id))
}

func TestGetAllFlights(t *testing.T) {
	store := db.NewInMemoryDB()
	require.Empty(t, slices.Collect(store.GetAllFlights()))

	flight1 := db.FlightPlan{
		Id:       uuid.New(),
		EntityID: scdussv1.EntityID(uuid.New().String()),
	}
	store.SaveFlight(flight1)
	saved := slices.Collect(store.GetAllFlights())
	require.Equal(t, []db.FlightPlan{flight1}, saved)

	flight2 := db.FlightPlan{
		Id:       uuid.New(),
		EntityID: scdussv1.EntityID(uuid.New().String()),
	}
	store.SaveFlight(flight2)
	saved = slices.Collect(store.GetAllFlights())
	require.ElementsMatch(t, []db.FlightPlan{flight1, flight2}, saved)
}
