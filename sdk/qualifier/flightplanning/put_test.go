package flightplanning_test

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
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

func putFlightPlan(
	t *testing.T,
	handler *flightplanning.Handler,
	flightID uuid.UUID,
	body flightplanning.UpsertFlightPlanRequest,
) *httptest.ResponseRecorder {
	t.Helper()
	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodPut, "/blah", bytes.NewReader(encoded))
	request.SetPathValue("flight_plan_id", flightID.String())
	recorder := httptest.NewRecorder()
	handler.PutFlightPlan(recorder, request)
	return recorder
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

func TestPutNewFlightPlanCreatesAnOperationalIntent(t *testing.T) {
	flightID := uuid.New()
	body := newFlightPlanBody()
	coordination, created := scdtest.NewCreateStub()
	handler := flightplanning.New(coordination, flightplanning.NewInMemoryFlightStore())

	response := putFlightPlan(t, handler, flightID, body)

	wiretest.RequireJSON(t, response, flightplanning.FlightPlanResponse{
		PlanningResult:   flightplanning.PlanningActivityResultCompleted,
		FlightPlanStatus: flightplanning.FlightPlanStatusPlanned,
	})
	require.Equal(t, scd.IntentParams{
		Volumes:  body.FlightPlan.BasicInformation.Area,
		State:    scdussv1.OperationalIntentState_Accepted,
		Priority: 2,
	}, created.Params)
	require.Equal(t,
		[]flightplanning.FlightPlanRecord{{ID: flightID, EntityID: created.EntityID}},
		handler.Flights.List())
}

func TestPutExistingFlightPlanUpdatesItsOperationalIntent(t *testing.T) {
	flightID := uuid.New()
	body := newFlightPlanBody()
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
			return scd.OperationalIntent{
				EntityID: id,
				State:    params.State,
			}, nil
		},
	}
	handler := flightplanning.New(coordination, flightplanning.NewInMemoryFlightStore())
	handler.Flights.Upsert(existing)

	response := putFlightPlan(t, handler, flightID, body)

	wiretest.RequireJSON(t, response, flightplanning.FlightPlanResponse{
		PlanningResult:   flightplanning.PlanningActivityResultCompleted,
		FlightPlanStatus: flightplanning.FlightPlanStatusPlanned,
	})
	require.Equal(t, existing.EntityID, receivedID)
	require.Equal(t, body.FlightPlan.BasicInformation.Area, received.Volumes)
	require.Equal(t, []flightplanning.FlightPlanRecord{existing}, handler.Flights.List())
}

func TestPutInUseFlightPlanActivatesIt(t *testing.T) {
	flightID := uuid.New()
	body := newFlightPlanBody()
	body.FlightPlan.BasicInformation.UsageState = flightplanning.UsageStateInUse
	coordination, created := scdtest.NewCreateStub()
	handler := flightplanning.New(coordination, flightplanning.NewInMemoryFlightStore())

	response := putFlightPlan(t, handler, flightID, body)

	wiretest.RequireJSON(t, response, flightplanning.FlightPlanResponse{
		PlanningResult:   flightplanning.PlanningActivityResultCompleted,
		FlightPlanStatus: flightplanning.FlightPlanStatusOkToFly,
	})
	require.Equal(t, scdussv1.OperationalIntentState_Activated, created.Params.State)
}

func TestPutOffNominalFlightPlanSubmitsNonconformingIntent(t *testing.T) {
	flightID := uuid.New()
	body := newFlightPlanBody()
	body.FlightPlan.BasicInformation.UsageState = flightplanning.UsageStateInUse
	body.FlightPlan.BasicInformation.UASState = flightplanning.UASStateOffNominal
	coordination, created := scdtest.NewCreateStub()
	handler := flightplanning.New(coordination, flightplanning.NewInMemoryFlightStore())

	response := putFlightPlan(t, handler, flightID, body)

	wiretest.RequireJSON(t, response, flightplanning.FlightPlanResponse{
		PlanningResult:   flightplanning.PlanningActivityResultCompleted,
		FlightPlanStatus: flightplanning.FlightPlanStatusOffNominal,
	})
	require.Equal(t, scdussv1.OperationalIntentState_Nonconforming, created.Params.State)
}

func TestPutRejectedFlightPlanIsNotPlanned(t *testing.T) {
	handler := flightplanning.New(rejectWith(scd.ErrRejected), flightplanning.NewInMemoryFlightStore())

	flightID := uuid.New()
	body := newFlightPlanBody()
	response := putFlightPlan(t, handler, flightID, body)

	wiretest.RequireJSON(t, response, flightplanning.FlightPlanResponse{
		PlanningResult:   flightplanning.PlanningActivityResultRejected,
		FlightPlanStatus: flightplanning.FlightPlanStatusNotPlanned,
	})
	require.Empty(t, handler.Flights.List())
}

func TestPutUnsupportedFlightPlanIsNotPlanned(t *testing.T) {
	handler := flightplanning.New(rejectWith(scd.ErrNotSupported), flightplanning.NewInMemoryFlightStore())

	flightID := uuid.New()
	body := newFlightPlanBody()
	response := putFlightPlan(t, handler, flightID, body)

	wiretest.RequireJSON(t, response, flightplanning.FlightPlanResponse{
		PlanningResult:   flightplanning.PlanningActivityResultNotSupported,
		FlightPlanStatus: flightplanning.FlightPlanStatusNotPlanned,
	})
	require.Empty(t, handler.Flights.List())
}

func TestPutNewFlightPlanInConflictIsNotPlanned(t *testing.T) {
	handler := flightplanning.New(rejectWith(scd.ErrConflict), flightplanning.NewInMemoryFlightStore())

	flightID := uuid.New()
	body := newFlightPlanBody()
	response := putFlightPlan(t, handler, flightID, body)

	wiretest.RequireJSON(t, response, flightplanning.FlightPlanResponse{
		PlanningResult:   flightplanning.PlanningActivityResultRejected,
		FlightPlanStatus: flightplanning.FlightPlanStatusNotPlanned,
	})
	require.Empty(t, handler.Flights.List())
}

func TestPutExistingFlightPlanInConflictStaysPlanned(t *testing.T) {
	coordination := scdtest.Stub{
		UpdateFn: func(context.Context, scdussv1.EntityID, scd.IntentParams) (scd.OperationalIntent, error) {
			return scd.OperationalIntent{
				EntityID: scdtest.NewEntityID(),
				State:    scdussv1.OperationalIntentState_Accepted,
			}, scd.ErrConflict
		},
	}
	handler := flightplanning.New(coordination, flightplanning.NewInMemoryFlightStore())
	flightID := uuid.New()
	body := newFlightPlanBody()
	existing := flightplanning.FlightPlanRecord{
		ID:       flightID,
		EntityID: scdtest.NewEntityID(),
	}
	handler.Flights.Upsert(existing)

	response := putFlightPlan(t, handler, flightID, body)

	wiretest.RequireJSON(t, response, flightplanning.FlightPlanResponse{
		PlanningResult:   flightplanning.PlanningActivityResultRejected,
		FlightPlanStatus: flightplanning.FlightPlanStatusPlanned,
	})
	require.ElementsMatch(t, []flightplanning.FlightPlanRecord{existing}, handler.Flights.List())
}
