package db

import (
	"slices"
	"testing"
	"uuid"

	"bwawan.com/openuss/internal/api/scdussv1"
	"github.com/stretchr/testify/require"
)

func TestGetIntentOnEmptyDB(t *testing.T) {
	db := NewInMemoryDB()
	id := scdussv1.EntityID(uuid.New().String())
	require.Nil(t, db.GetIntent(id))
}

func TestDeleteIntentOnEmptyDB(t *testing.T) {
	db := NewInMemoryDB()
	id := scdussv1.EntityID(uuid.New().String())
	require.NotPanics(t, func() {
		db.DeleteIntent(id)
	})
}

func TestGetSavedIntent(t *testing.T) {
	db := NewInMemoryDB()
	intent := OperationalIntent{
		EntityID: scdussv1.EntityID(uuid.New().String()),
	}
	require.NoError(t, db.SaveIntent(intent))
	require.Equal(t, intent, *(db.GetIntent(intent.EntityID)))
}

func TestDeleteSavedIntent(t *testing.T) {
	db := NewInMemoryDB()
	id := scdussv1.EntityID(uuid.New().String())
	intent := OperationalIntent{
		EntityID: id,
	}
	require.NoError(t, db.SaveIntent(intent))
	require.Equal(t, intent, *db.GetIntent(id))
	db.DeleteIntent(id)
	require.Nil(t, db.GetIntent(id))
}

func TestGetAllIntents(t *testing.T) {
	db := NewInMemoryDB()
	require.Empty(t, slices.Collect(db.GetAllIntents()))

	intent1 := OperationalIntent{
		EntityID: scdussv1.EntityID(uuid.New().String()),
	}
	db.SaveIntent(intent1)
	saved := slices.Collect(db.GetAllIntents())
	require.Equal(t, []OperationalIntent{intent1}, saved)

	intent2 := OperationalIntent{
		EntityID: scdussv1.EntityID(uuid.New().String()),
	}
	db.SaveIntent(intent2)
	saved = slices.Collect(db.GetAllIntents())
	require.ElementsMatch(t, []OperationalIntent{intent1, intent2}, saved)
}

func TestGetFlightOnEmptyDB(t *testing.T) {
	db := NewInMemoryDB()
	require.Nil(t, db.GetFlight(uuid.New()))
}

func TestDeleteFlightOnEmptyDB(t *testing.T) {
	db := NewInMemoryDB()
	require.NotPanics(t, func() {
		db.DeleteFlight(uuid.New())
	})
}

func TestGetSavedFlight(t *testing.T) {
	db := NewInMemoryDB()
	flight := FlightPlan{
		Id: uuid.New(),
	}
	require.NoError(t, db.SaveFlight(flight))
	require.Equal(t, flight, *(db.GetFlight(flight.Id)))
}

func TestDeleteSavedFlight(t *testing.T) {
	db := NewInMemoryDB()
	flight := FlightPlan{
		Id: uuid.New(),
	}
	require.NoError(t, db.SaveFlight(flight))
	require.Equal(t, flight, *db.GetFlight(flight.Id))
	db.DeleteFlight(flight.Id)
	require.Nil(t, db.GetFlight(flight.Id))
}

func TestGetAllFlights(t *testing.T) {
	db := NewInMemoryDB()
	require.Empty(t, slices.Collect(db.GetAllFlights()))

	flight1 := FlightPlan{
		Id:       uuid.New(),
		EntityID: scdussv1.EntityID(uuid.New().String()),
	}
	db.SaveFlight(flight1)
	saved := slices.Collect(db.GetAllFlights())
	require.Equal(t, []FlightPlan{flight1}, saved)

	flight2 := FlightPlan{
		Id:       uuid.New(),
		EntityID: scdussv1.EntityID(uuid.New().String()),
	}
	db.SaveFlight(flight2)
	saved = slices.Collect(db.GetAllFlights())
	require.ElementsMatch(t, []FlightPlan{flight1, flight2}, saved)
}
