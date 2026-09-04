package operations

import (
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/db"
	"github.com/stretchr/testify/require"
)

func newHandler() *Handler {
	db := db.NewInMemoryDB()
	return &Handler{DB: db}
}

func TestGetOperationalIntentMissingEntityId(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/foo", nil)
	handler := newHandler()
	handler.GetOperationalIntent(recorder, request)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGetOperationalIntentNotExists(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/foo", nil)
	request.SetPathValue("entity_id", uuid.New().String())
	handler := newHandler()
	handler.GetOperationalIntent(recorder, request)
	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestGetOperationalIntentSuccess(t *testing.T) {
	entityId := uuid.New().String()
	intent := scdussv1.OperationalIntent{
		Reference: scdussv1.OperationalIntentReference{
			Id: scdussv1.EntityID(entityId),
		},
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/foo", nil)
	request.SetPathValue("entity_id", entityId)
	handler := newHandler()
	handler.DB.SaveIntent(intent)

	handler.GetOperationalIntent(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var result scdussv1.GetOperationalIntentDetailsResponse
	err := json.UnmarshalRead(recorder.Body, &result)
	require.NoError(t, err)
	require.Equal(t, intent, result.OperationalIntent)
}
