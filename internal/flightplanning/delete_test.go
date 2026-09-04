package flightplanning

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/db"
	"bwawan.com/openuss/internal/testutil"
	"github.com/stretchr/testify/require"
)

func newDeleteRequest(flightPlanId *string) *http.Request {
	request := httptest.NewRequest(http.MethodDelete, "/blah", nil)
	if flightPlanId != nil {
		request.SetPathValue("flight_plan_id", *flightPlanId)
	}
	return request
}

func newFlightPlan(t *testing.T, handler *Handler) db.FlightPlan {
	id := scdussv1.EntityID(uuid.New().String())
	params := scdussv1.PutOperationalIntentReferenceParameters{
		Extents: testutil.NewVolumes4D(),
	}
	result, _ := handler.DSS.CreateOperationalIntentReference(t.Context(), id, params)
	reference := result.OperationalIntentReference
	intent := db.OperationalIntent{
		EntityID: reference.Id,
		Ovn:      *reference.Ovn,
	}
	flight := db.FlightPlan{
		Id:       uuid.New(),
		EntityID: intent.EntityID,
	}
	handler.DB.SaveIntent(intent)
	handler.DB.SaveFlight(flight)
	return flight
}

func TestDeleteFlightPlanSucceeds(t *testing.T) {
	handler, dss := newHandler()
	flight := newFlightPlan(t, handler)

	recorder := httptest.NewRecorder()
	request := newDeleteRequest(new(flight.Id.String()))
	handler.DeleteFlightPlan(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Nil(t, handler.DB.GetFlight(flight.Id))
	require.Nil(t, handler.DB.GetIntent(flight.EntityID))
	require.NotContains(t, dss.Intents, flight.EntityID)
	testutil.RequireJSON(t, recorder, map[string]any{
		"flight_plan_status": "Closed",
		"planning_result":    "Completed",
	})
}

func TestDeleteFlightPlanMissingFlightId(t *testing.T) {
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

func TestDeleteFlightPlanFails(t *testing.T) {
	handler, dss := newHandler()
	flight := newFlightPlan(t, handler)
	intent := handler.DB.GetIntent(flight.EntityID)
	intent.Ovn = scdussv1.EntityOVN(uuid.New().String())
	handler.DB.SaveIntent(*intent)

	recorder := httptest.NewRecorder()
	request := newDeleteRequest(new(flight.Id.String()))
	handler.DeleteFlightPlan(recorder, request)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Equal(t, flight, *handler.DB.GetFlight(flight.Id))
	require.Equal(t, intent, handler.DB.GetIntent(flight.EntityID))
	require.Contains(t, dss.Intents, flight.EntityID)
}
