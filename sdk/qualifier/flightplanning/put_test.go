package flightplanning_test

import (
	"context"
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/internal/wiretest"
	"bwawan.com/openuss/sdk/qualifier/flightplanning"
	"bwawan.com/openuss/sdk/scd"
	"bwawan.com/openuss/sdk/scdtest"
	"github.com/stretchr/testify/require"
)

func newFlightPlanBody() flightplanning.UpsertFlightPlanRequest {
	return flightplanning.UpsertFlightPlanRequest{
		RequestID: scdussv1.UUIDv4Format(uuid.New().String()),
		FlightPlan: flightplanning.FlightPlan{
			BasicInformation: flightplanning.BasicFlightPlanInformation{
				Area: scdtest.NewVolumes4D(),
			},
			F3548: flightplanning.ASTMF354821OpIntentInformation{
				Priority: 2,
			},
		},
	}
}

func newFlightParams() (uuid.UUID, flightplanning.UpsertFlightPlanRequest) {
	return uuid.New(), newFlightPlanBody()
}

func putFlightPlanRequest(id *uuid.UUID, body flightplanning.UpsertFlightPlanRequest) *http.Request {
	bytes, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	request := httptest.NewRequest(http.MethodPut, "/blah", strings.NewReader(string(bytes)))
	if id != nil {
		request.SetPathValue("flight_plan_id", id.String())
	}
	return request
}

func rejectWith(err error) scdtest.Stub {
	return scdtest.Stub{
		CreateFn: func(context.Context, scd.IntentParams) (scd.OperationalIntent, error) {
			return scd.OperationalIntent{}, err
		},
		UpdateFn: func(context.Context, scdussv1.EntityID, scd.IntentParams) (scd.OperationalIntent, error) {
			return scd.OperationalIntent{}, err
		},
	}
}

func TestCreateFlightPlanSucceeds(t *testing.T) {
	flightID, flight := newFlightParams()
	intent := scd.OperationalIntent{EntityID: scdtest.NewEntityID()}
	var received scd.IntentParams
	coordination := scdtest.Stub{
		CreateFn: func(_ context.Context, params scd.IntentParams) (scd.OperationalIntent, error) {
			received = params
			return intent, nil
		},
	}
	handler := flightplanning.New(coordination, flightplanning.NewInMemoryFlightStore())

	response := httptest.NewRecorder()
	handler.PutFlightPlan(response, putFlightPlanRequest(&flightID, flight))

	wiretest.RequireJSON(t, response, flightplanning.FlightPlanResponse{
		PlanningResult:   flightplanning.PlanningActivityResultCompleted,
		FlightPlanStatus: flightplanning.FlightPlanStatusPlanned,
	})
	require.Equal(t, scd.IntentParams{
		Volumes:  flight.FlightPlan.BasicInformation.Area,
		State:    scdussv1.OperationalIntentState_Accepted,
		Priority: 2,
	}, received)
	require.Equal(t,
		[]flightplanning.FlightPlanRecord{{ID: flightID, EntityID: intent.EntityID}},
		handler.Flights.List())
}

func TestUpdateFlightPlanSucceeds(t *testing.T) {
	flightID, flight := newFlightParams()
	existing := flightplanning.FlightPlanRecord{ID: flightID, EntityID: scdtest.NewEntityID()}
	var receivedID scdussv1.EntityID
	var received scd.IntentParams
	coordination := scdtest.Stub{
		UpdateFn: func(
			_ context.Context,
			id scdussv1.EntityID,
			params scd.IntentParams,
		) (scd.OperationalIntent, error) {
			receivedID, received = id, params
			return scd.OperationalIntent{EntityID: id}, nil
		},
	}
	handler := flightplanning.New(coordination, flightplanning.NewInMemoryFlightStore())
	handler.Flights.Upsert(existing)

	response := httptest.NewRecorder()
	handler.PutFlightPlan(response, putFlightPlanRequest(&flightID, flight))

	wiretest.RequireJSON(t, response, flightplanning.FlightPlanResponse{
		PlanningResult:   flightplanning.PlanningActivityResultCompleted,
		FlightPlanStatus: flightplanning.FlightPlanStatusOkToFly,
	})
	require.Equal(t, existing.EntityID, receivedID)
	require.Equal(t, flight.FlightPlan.BasicInformation.Area, received.Volumes)
	require.Equal(t, []flightplanning.FlightPlanRecord{existing}, handler.Flights.List())
}

func TestUsageStateInUseActivatesFlightPlan(t *testing.T) {
	flightID, flight := newFlightParams()
	flight.FlightPlan.BasicInformation.UsageState = flightplanning.UsageStateInUse
	var received scd.IntentParams
	coordination := scdtest.Stub{
		CreateFn: func(_ context.Context, params scd.IntentParams) (scd.OperationalIntent, error) {
			received = params
			return scd.OperationalIntent{EntityID: scdtest.NewEntityID()}, nil
		},
	}
	handler := flightplanning.New(coordination, flightplanning.NewInMemoryFlightStore())

	response := httptest.NewRecorder()
	handler.PutFlightPlan(response, putFlightPlanRequest(&flightID, flight))

	wiretest.RequireJSON(t, response, flightplanning.FlightPlanResponse{
		PlanningResult:   flightplanning.PlanningActivityResultCompleted,
		FlightPlanStatus: flightplanning.FlightPlanStatusPlanned,
	})
	require.Equal(t, scdussv1.OperationalIntentState_Activated, received.State)
}

func TestPutRejectedFlightPlanIsNotPlanned(t *testing.T) {
	handler := flightplanning.New(rejectWith(scd.ErrRejected), flightplanning.NewInMemoryFlightStore())

	flightID, flight := newFlightParams()
	response := httptest.NewRecorder()
	handler.PutFlightPlan(response, putFlightPlanRequest(&flightID, flight))

	wiretest.RequireJSON(t, response, flightplanning.FlightPlanResponse{
		PlanningResult:   flightplanning.PlanningActivityResultRejected,
		FlightPlanStatus: flightplanning.FlightPlanStatusNotPlanned,
	})
	require.Empty(t, handler.Flights.List())
}

func TestPutNewFlightPlanInConflictIsNotPlanned(t *testing.T) {
	handler := flightplanning.New(rejectWith(scd.ErrConflict), flightplanning.NewInMemoryFlightStore())

	flightID, flight := newFlightParams()
	response := httptest.NewRecorder()
	handler.PutFlightPlan(response, putFlightPlanRequest(&flightID, flight))

	wiretest.RequireJSON(t, response, flightplanning.FlightPlanResponse{
		PlanningResult:   flightplanning.PlanningActivityResultRejected,
		FlightPlanStatus: flightplanning.FlightPlanStatusNotPlanned,
	})
	require.Empty(t, handler.Flights.List())
}

func TestPutExistingFlightPlanInConflictStaysPlanned(t *testing.T) {
	handler := flightplanning.New(rejectWith(scd.ErrConflict), flightplanning.NewInMemoryFlightStore())

	flightID, flightParams := newFlightParams()
	flight := flightplanning.FlightPlanRecord{
		ID:       flightID,
		EntityID: scdtest.NewEntityID(),
	}
	handler.Flights.Upsert(flight)

	response := httptest.NewRecorder()
	handler.PutFlightPlan(response, putFlightPlanRequest(&flightID, flightParams))

	wiretest.RequireJSON(t, response, flightplanning.FlightPlanResponse{
		PlanningResult:   flightplanning.PlanningActivityResultRejected,
		FlightPlanStatus: flightplanning.FlightPlanStatusPlanned,
	})
	require.ElementsMatch(t, []flightplanning.FlightPlanRecord{flight}, handler.Flights.List())
}
