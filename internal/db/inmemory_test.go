package db

import (
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
