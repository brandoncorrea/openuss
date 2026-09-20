package flightplanning_test

import (
	"encoding/json/v2"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"
	"uuid"

	"bwawan.com/openuss/internal/db"
	"bwawan.com/openuss/internal/flightplanning"
	"bwawan.com/openuss/internal/wiretest"
	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/auth"
	"bwawan.com/openuss/sdk/dss"
	"bwawan.com/openuss/sdk/dss/dsstest"
	"bwawan.com/openuss/sdk/peer"
	"bwawan.com/openuss/sdk/scd"
	"bwawan.com/openuss/sdk/scdtest"
	"bwawan.com/openuss/sdk/utmclient"
	"github.com/stretchr/testify/require"
)

func newFlightPlanBody() flightplanning.PutFlightPlanBody {
	return flightplanning.PutFlightPlanBody{
		RequestID: scdussv1.UUIDv4Format(uuid.New().String()),
		FlightPlan: flightplanning.FlightPlan{
			BasicInformation: flightplanning.FlightPlanBasicInformation{
				Area: scdtest.NewVolumes4D(),
			},
			F3548: flightplanning.F3548{
				Priority: 2,
			},
		},
	}
}

func newFlightParams() (uuid.UUID, flightplanning.PutFlightPlanBody) {
	return uuid.New(), newFlightPlanBody()
}

func putFlightPlanRequest(id *uuid.UUID, body flightplanning.PutFlightPlanBody) *http.Request {
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

func newHandler() (*flightplanning.Handler, *dss.InMemoryDSS) {
	dssClient := dss.NewInMemoryDSS()
	service := scd.New(dssClient, nil, scd.NewInMemoryIntentStore(), "http://openuss.localutm")
	return flightplanning.New(service, db.NewInMemoryDB()), dssClient
}

func TestCreateFlightPlanSucceeds(t *testing.T) {
	flightID, flight := newFlightParams()
	response := httptest.NewRecorder()
	request := putFlightPlanRequest(&flightID, flight)
	handler, _ := newHandler()
	handler.PutFlightPlan(response, request)

	wiretest.RequireJSON(t, response, map[string]any{
		"planning_result":    "Completed",
		"flight_plan_status": "Planned",
	})

	require.Len(t, slices.Collect(handler.DB.GetAllFlights()), 1)
	intent, err := handler.SCD.Intents.Get(t.Context(), handler.DB.GetFlight(flightID).EntityID)
	require.NoError(t, err)
	require.Equal(t, flight.FlightPlan.BasicInformation.Area, intent.Volumes)
	require.EqualValues(t, 2, intent.Priority)
	require.Equal(t, scdussv1.OperationalIntentState_Accepted, intent.State)
}

func TestUpdateFlightPlanSucceeds(t *testing.T) {
	flightID, flight := newFlightParams()
	createResponse := httptest.NewRecorder()
	handler, _ := newHandler()
	handler.PutFlightPlan(createResponse, putFlightPlanRequest(&flightID, flight))

	wiretest.RequireJSON(t, createResponse, map[string]any{
		"planning_result":    "Completed",
		"flight_plan_status": "Planned",
	})

	flight1 := handler.DB.GetFlight(flightID)

	updateResponse := httptest.NewRecorder()
	flight.FlightPlan.BasicInformation.Area[0].Volume.AltitudeLower.Value += 1
	handler.PutFlightPlan(updateResponse, putFlightPlanRequest(&flightID, flight))

	wiretest.RequireJSON(t, updateResponse, map[string]any{
		"planning_result":    "Completed",
		"flight_plan_status": "OkToFly",
	})

	require.Len(t, slices.Collect(handler.DB.GetAllFlights()), 1)
	require.Equal(t, flight1, handler.DB.GetFlight(flightID))

	intents, err := handler.SCD.Intents.List(t.Context())
	require.NoError(t, err)
	require.Len(t, intents, 1)
	require.Equal(t, flight1.EntityID, intents[0].EntityID)
	require.Equal(t, flight.FlightPlan.BasicInformation.Area, intents[0].Volumes)
}

func TestPutRejectedFlightPlanIsNotPlanned(t *testing.T) {
	flightID, flight := newFlightParams()
	response := httptest.NewRecorder()
	tooLate := time.Now().Add(time.Hour * 24 * 30).Add(time.Second)
	area := flight.FlightPlan.BasicInformation.Area[0]
	area.TimeStart.Value = tooLate.Format(time.RFC3339Nano)
	area.TimeEnd.Value = tooLate.Add(time.Hour).Format(time.RFC3339Nano)
	planner, _ := newHandler()
	planner.PutFlightPlan(response, putFlightPlanRequest(&flightID, flight))
	wiretest.RequireJSON(t, response, map[string]any{
		"activity_result":    "Rejected",
		"planning_result":    "Rejected",
		"flight_plan_status": "NotPlanned",
	})
}

func TestUsageStateInUseActivatesFlightPlan(t *testing.T) {
	response := httptest.NewRecorder()
	flightID, flight := newFlightParams()
	flight.FlightPlan.BasicInformation.UsageState = "InUse"
	request := putFlightPlanRequest(&flightID, flight)
	handler, dssClient := newHandler()
	handler.PutFlightPlan(response, request)

	wiretest.RequireJSON(t, response, map[string]any{
		"planning_result":    "Completed",
		"flight_plan_status": "Planned",
	})

	dssIntent := slices.Collect(maps.Values(dssClient.Intents))[0]
	require.Equal(t, scdussv1.OperationalIntentState_Activated, dssIntent.Reference.State)
}

func newPriority100Peer() []scdussv1.OperationalIntent {
	return []scdussv1.OperationalIntent{
		{
			Reference: scdussv1.OperationalIntentReference{
				Id:         scdussv1.EntityID(uuid.New().String()),
				Ovn:        new(scdussv1.EntityOVN(uuid.New().String())),
				UssBaseUrl: scdussv1.OperationalIntentUssBaseURL("http://uss1.localutm"),
			},
			Details: scdussv1.OperationalIntentDetails{
				Priority: new(scdussv1.Priority(100)),
			},
		},
	}
}

func TestPutNewFlightPlanInConflictIsNotPlanned(t *testing.T) {
	handler := newHandlerFromPeers(t, newPriority100Peer())

	flightID, flight := newFlightParams()
	response := httptest.NewRecorder()
	request := putFlightPlanRequest(&flightID, flight)
	handler.PutFlightPlan(response, request)

	wiretest.RequireJSON(t, response, map[string]any{
		"activity_result":    "Rejected",
		"planning_result":    "Rejected",
		"flight_plan_status": "NotPlanned",
	})
	require.Empty(t, slices.Collect(handler.DB.GetAllFlights()))
}

func TestPutExistingFlightPlanInConflictStaysPlanned(t *testing.T) {
	handler := newHandlerFromPeers(t, newPriority100Peer())

	flightID, flightParams := newFlightParams()

	intent := scd.OperationalIntent{
		EntityID: scdussv1.EntityID(uuid.New().String()),
		OVN:      scdussv1.EntityOVN(uuid.New().String()),
	}
	flight := db.FlightPlan{
		ID:       flightID,
		EntityID: intent.EntityID,
	}
	require.NoError(t, handler.SCD.Intents.Upsert(t.Context(), intent))
	handler.DB.SaveFlight(flight)

	response := httptest.NewRecorder()
	request := putFlightPlanRequest(&flightID, flightParams)
	handler.PutFlightPlan(response, request)

	wiretest.RequireJSON(t, response, map[string]any{
		"activity_result":    "Rejected",
		"planning_result":    "Rejected",
		"flight_plan_status": "Planned",
	})
	require.ElementsMatch(t, []db.FlightPlan{flight}, slices.Collect(handler.DB.GetAllFlights()))
}

func newHandlerFromPeers(t *testing.T, peers []scdussv1.OperationalIntent) *flightplanning.Handler {
	client := utmclient.New(
		auth.NewInMemoryTokenSource(),
		httptest.NewTestServer(t, dsstest.NewPeerHandler(peers)).Client(),
	)
	service := scd.New(
		dss.New("https://dss.localutm", client),
		peer.New(client),
		scd.NewInMemoryIntentStore(),
		"http://openuss.localutm",
	)
	return flightplanning.New(service, db.NewInMemoryDB())
}
