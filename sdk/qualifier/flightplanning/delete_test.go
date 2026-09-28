package flightplanning_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/internal/wiretest"
	"bwawan.com/openuss/sdk/qualifier/flightplanning"
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

	wiretest.RequireJSON(t, recorder, flightplanning.FlightPlanResponse{
		FlightPlanStatus: flightplanning.FlightPlanStatusClosed,
		PlanningResult:   flightplanning.PlanningActivityResultCompleted,
	})
	require.Equal(t, flight.EntityID, deleted)
	require.Nil(t, handler.Flights.Get(flight.ID))
}
