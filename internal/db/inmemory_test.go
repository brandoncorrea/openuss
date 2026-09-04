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
	intent := scdussv1.OperationalIntent{}
	require.NoError(t, db.SaveIntent(intent))
	require.Equal(t, intent, *(db.GetIntent(intent.Reference.Id)))
}

func TestDeleteSavedIntent(t *testing.T) {
	db := NewInMemoryDB()
	id := scdussv1.EntityID(uuid.New().String())
	intent := scdussv1.OperationalIntent{
		Reference: scdussv1.OperationalIntentReference{
			Id: id,
		},
	}
	require.NoError(t, db.SaveIntent(intent))
	require.Equal(t, intent, *db.GetIntent(id))
	db.DeleteIntent(id)
	require.Nil(t, db.GetIntent(id))
}
