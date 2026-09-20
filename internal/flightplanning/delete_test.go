package flightplanning_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"bwawan.com/openuss/internal/db"
	"bwawan.com/openuss/internal/flightplanning"
	"bwawan.com/openuss/internal/wiretest"
	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/scd"
	"bwawan.com/openuss/sdk/scdtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDeleteRequest(flightPlanID *string) *http.Request {
	request := httptest.NewRequest(http.MethodDelete, "/blah", nil)
	if flightPlanID != nil {
		request.SetPathValue("flight_plan_id", *flightPlanID)
	}
	return request
}

func newFlightPlan(t *testing.T, handler *flightplanning.Handler) db.FlightPlan {
	id := scdtest.NewEntityID()
	params := scdussv1.PutOperationalIntentReferenceParameters{
		Extents: scdtest.NewVolumes4D(),
	}
	result, _ := scdService(t, handler).DSS.PutOperationalIntentReference(t.Context(), id, nil, params)
	reference := result.OperationalIntentReference
	intent := scd.OperationalIntent{
		EntityID: reference.Id,
		OVN:      *reference.Ovn,
	}
	flight := db.FlightPlan{
		ID:       uuid.New(),
		EntityID: intent.EntityID,
	}
	require.NoError(t, scdService(t, handler).Intents.Upsert(t.Context(), intent))
	handler.DB.SaveFlight(flight)
	return flight
}

func TestDeleteFlightPlanSucceeds(t *testing.T) {
	handler, dssClient := newHandler()
	flight := newFlightPlan(t, handler)

	recorder := httptest.NewRecorder()
	request := newDeleteRequest(new(flight.ID.String()))
	handler.DeleteFlightPlan(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Nil(t, handler.DB.GetFlight(flight.ID))

	_, err := scdService(t, handler).Intents.Get(t.Context(), flight.EntityID)
	require.ErrorIs(t, err, scd.ErrNotFound)

	_, registered := dssClient.OperationalIntent(flight.EntityID)
	require.False(t, registered)
	wiretest.RequireJSON(t, recorder, map[string]any{
		"flight_plan_status": "Closed",
		"planning_result":    "Completed",
	})
}

func TestDeleteFlightPlanMissingFlightID(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := newDeleteRequest(nil)
	handler, _ := newHandler()
	handler.DeleteFlightPlan(recorder, request)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDeleteFlightPlanDoesNotExist(t *testing.T) {
	recorder := httptest.NewRecorder()
	id := new(uuid.New().String())
	request := newDeleteRequest(id)
	request.SetPathValue("flight_plan_id", uuid.New().String())
	handler, _ := newHandler()
	handler.DeleteFlightPlan(recorder, request)
	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestDeleteFlightPlanFailsWhenCoordinationFails(t *testing.T) {
	flight := db.FlightPlan{
		ID:       uuid.New(),
		EntityID: scdtest.NewEntityID(),
	}
	unavailable := scdtest.Stub{
		DeleteFn: func(_ context.Context, id scdussv1.EntityID) error {
			assert.Equal(t, flight.EntityID, id)
			return errors.New("dss unavailable")
		},
	}
	handler := flightplanning.New(unavailable, db.NewInMemoryDB())
	handler.DB.SaveFlight(flight)

	recorder := httptest.NewRecorder()
	handler.DeleteFlightPlan(recorder, newDeleteRequest(new(flight.ID.String())))

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Equal(t, flight, *handler.DB.GetFlight(flight.ID))
}
