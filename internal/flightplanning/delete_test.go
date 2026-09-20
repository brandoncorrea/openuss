package flightplanning_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"bwawan.com/openuss/internal/flightplanning"
	"bwawan.com/openuss/internal/wiretest"
	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/scdtest"
	"github.com/stretchr/testify/require"
)

func newDeleteRequest(flightPlanID *string) *http.Request {
	request := httptest.NewRequest(http.MethodDelete, "/blah", nil)
	if flightPlanID != nil {
		request.SetPathValue("flight_plan_id", *flightPlanID)
	}
	return request
}

func newSavedFlight(handler *flightplanning.Handler) flightplanning.FlightPlanRecord {
	flight := flightplanning.FlightPlanRecord{
		ID:       uuid.New(),
		EntityID: scdtest.NewEntityID(),
	}
	handler.Flights.Upsert(flight)
	return flight
}

func TestDeleteFlightPlanSucceeds(t *testing.T) {
	var deleted scdussv1.EntityID
	coordination := scdtest.Stub{
		DeleteFn: func(_ context.Context, id scdussv1.EntityID) error {
			deleted = id
			return nil
		},
	}
	handler := flightplanning.New(coordination, flightplanning.NewInMemoryFlightStore())
	flight := newSavedFlight(handler)

	recorder := httptest.NewRecorder()
	handler.DeleteFlightPlan(recorder, newDeleteRequest(new(flight.ID.String())))

	wiretest.RequireJSON(t, recorder, map[string]any{
		"flight_plan_status": "Closed",
		"planning_result":    "Completed",
	})
	require.Equal(t, flight.EntityID, deleted)
	require.Nil(t, handler.Flights.Get(flight.ID))
}

func TestDeleteFlightPlanMissingFlightID(t *testing.T) {
	handler := flightplanning.New(scdtest.Stub{}, flightplanning.NewInMemoryFlightStore())

	recorder := httptest.NewRecorder()
	handler.DeleteFlightPlan(recorder, newDeleteRequest(nil))

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDeleteFlightPlanDoesNotExist(t *testing.T) {
	handler := flightplanning.New(scdtest.Stub{}, flightplanning.NewInMemoryFlightStore())

	recorder := httptest.NewRecorder()
	handler.DeleteFlightPlan(recorder, newDeleteRequest(new(uuid.New().String())))

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestDeleteFlightPlanFailsWhenCoordinationFails(t *testing.T) {
	unavailable := scdtest.Stub{
		DeleteFn: func(context.Context, scdussv1.EntityID) error {
			return errors.New("dss unavailable")
		},
	}
	handler := flightplanning.New(unavailable, flightplanning.NewInMemoryFlightStore())
	flight := newSavedFlight(handler)

	recorder := httptest.NewRecorder()
	handler.DeleteFlightPlan(recorder, newDeleteRequest(new(flight.ID.String())))

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Equal(t, flight, *handler.Flights.Get(flight.ID))
}
