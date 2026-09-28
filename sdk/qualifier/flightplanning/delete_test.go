package flightplanning_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"bwawan.com/openuss/sdk/internal/wiretest"
	"bwawan.com/openuss/sdk/qualifier/flightplanning"
	"bwawan.com/openuss/sdk/scdtest"
	"github.com/stretchr/testify/require"
)

func newDeleteRequest(flightID uuid.UUID) *http.Request {
	request := httptest.NewRequest(http.MethodDelete, "/blah", nil)
	request.SetPathValue("flight_plan_id", flightID.String())
	return request
}

func newSavedRecord(handler *flightplanning.Handler) flightplanning.FlightPlanRecord {
	record := flightplanning.FlightPlanRecord{
		ID:       uuid.New(),
		EntityID: scdtest.NewEntityID(),
	}
	handler.Flights.Upsert(record)
	return record
}

func TestDeleteFlightPlanSucceeds(t *testing.T) {
	coordination, deleted := scdtest.NewDeleteStub()
	handler := flightplanning.New(coordination, flightplanning.NewInMemoryFlightStore())
	record := newSavedRecord(handler)

	recorder := httptest.NewRecorder()
	handler.DeleteFlightPlan(recorder, newDeleteRequest(record.ID))

	wiretest.RequireJSON(t, recorder, flightplanning.FlightPlanResponse{
		FlightPlanStatus: flightplanning.FlightPlanStatusClosed,
		PlanningResult:   flightplanning.PlanningActivityResultCompleted,
	})
	require.Equal(t, record.EntityID, deleted.EntityID)
	require.Nil(t, handler.Flights.Get(record.ID))
}
