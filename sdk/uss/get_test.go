package uss_test

import (
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"uuid"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/scd"
	"bwawan.com/openuss/sdk/scdtest"
	"bwawan.com/openuss/sdk/uss"
	"github.com/stretchr/testify/require"
)

func newHandler() *uss.Handler {
	return uss.New(scd.NewInMemoryIntentStore())
}

func TestGetOperationalIntentMissingEntityID(t *testing.T) {
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
	entityID := uuid.New().String()
	intent := scd.OperationalIntent{
		EntityID:        scdussv1.EntityID(entityID),
		Manager:         "the-manager",
		USSAvailability: scdussv1.UssAvailabilityState_Normal,
		Version:         1,
		Priority:        2,
		State:           scdussv1.OperationalIntentState_Accepted,
		OVN:             scdussv1.EntityOVN(uuid.New().String()),
		TimeStart:       time.Now().Add(time.Minute),
		TimeEnd:         time.Now().Add(6 * time.Minute),
		USSBaseURL:      "the-uss-base-url",
		SubscriptionID:  scdussv1.SubscriptionID(uuid.New().String()),
		Volumes:         scdtest.NewVolumes4D(),
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/foo", nil)
	request.SetPathValue("entity_id", entityID)
	handler := newHandler()
	require.NoError(t, handler.Intents.Upsert(t.Context(), intent))

	handler.GetOperationalIntent(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var result scdussv1.GetOperationalIntentDetailsResponse
	err := json.UnmarshalRead(recorder.Body, &result)
	require.NoError(t, err)
	require.Equal(t, intent.EntityID, result.OperationalIntent.Reference.Id)
	require.Equal(t, "the-manager", result.OperationalIntent.Reference.Manager)
	require.Equal(t, scdussv1.UssAvailabilityState_Normal, result.OperationalIntent.Reference.UssAvailability)
	require.EqualValues(t, 1, result.OperationalIntent.Reference.Version)
	require.Equal(t, scdussv1.OperationalIntentState_Accepted, result.OperationalIntent.Reference.State)
	require.Equal(t, "RFC3339", result.OperationalIntent.Reference.TimeStart.Format)
	require.Equal(t, intent.TimeStart.Format(time.RFC3339Nano), result.OperationalIntent.Reference.TimeStart.Value)
	require.Equal(t, "RFC3339", result.OperationalIntent.Reference.TimeEnd.Format)
	require.Equal(t, intent.TimeEnd.Format(time.RFC3339Nano), result.OperationalIntent.Reference.TimeEnd.Value)
	require.EqualValues(t, "the-uss-base-url", result.OperationalIntent.Reference.UssBaseUrl)
	require.Equal(t, intent.SubscriptionID, result.OperationalIntent.Reference.SubscriptionId)
	require.Equal(t, intent.Volumes, *result.OperationalIntent.Details.Volumes)
	require.EqualValues(t, 2, *result.OperationalIntent.Details.Priority)
}
