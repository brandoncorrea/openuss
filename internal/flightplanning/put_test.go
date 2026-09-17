package flightplanning

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

	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/auth"
	"bwawan.com/openuss/internal/db"
	"bwawan.com/openuss/internal/dss"
	"bwawan.com/openuss/internal/peer"
	"bwawan.com/openuss/internal/testutil"
	"bwawan.com/openuss/internal/utmclient"
	"github.com/stretchr/testify/require"
)

func newFlightPlanBody() PutFlightPlanBody {
	return PutFlightPlanBody{
		RequestId: scdussv1.UUIDv4Format(uuid.New().String()),
		FlightPlan: FlightPlan{
			BasicInformation: FlightPlanBasicInformation{
				Area: testutil.NewVolumes4D(),
			},
			Astm: AstmF3548v21{
				Priority: 2,
			},
		},
	}
}

func newFlightParams() (uuid.UUID, PutFlightPlanBody) {
	return uuid.New(), newFlightPlanBody()
}

func putFlightPlanRequest(id *uuid.UUID, body PutFlightPlanBody) *http.Request {
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

func newHandler() (*Handler, *dss.InMemoryDSS) {
	dss := dss.NewInMemoryDSS()
	db := db.NewInMemoryDB()
	handler := Handler{
		DSS:        dss,
		DB:         db,
		UssBaseUrl: scdussv1.OperationalIntentUssBaseURL("http://openuss.localutm"),
	}
	return &handler, dss
}

func TestCreateFlightPlanSucceeds(t *testing.T) {
	flightId, flight := newFlightParams()
	response := httptest.NewRecorder()
	request := putFlightPlanRequest(&flightId, flight)
	handler, dss := newHandler()
	handler.PutFlightPlan(response, request)

	testutil.RequireJSON(t, response, map[string]any{
		"planning_result":    "Completed",
		"flight_plan_status": "Planned",
	})

	memoryDb := handler.DB.(*db.InMemoryDB)
	require.Len(t, dss.Intents, 1)

	dssIntent := slices.Collect(maps.Values(dss.Intents))[0]
	require.Equal(t, flight.FlightPlan.BasicInformation.Area, *dssIntent.Details.Volumes)
	require.Equal(t, scdussv1.OperationalIntentState_Accepted, dssIntent.Reference.State)
	require.EqualValues(t, "http://openuss.localutm", dssIntent.Reference.UssBaseUrl)

	savedIntent := memoryDb.GetIntent(dssIntent.Reference.Id)
	require.Equal(t, dssIntent.Reference.Id, savedIntent.EntityID)
	require.Equal(t, "InMemoryManager", savedIntent.Manager)
	require.Equal(t, scdussv1.UssAvailabilityState_Normal, savedIntent.UssAvailability)
	require.EqualValues(t, 1, savedIntent.Version)
	require.EqualValues(t, 2, savedIntent.Priority)
	require.Equal(t, scdussv1.OperationalIntentState_Accepted, savedIntent.State)
	require.Equal(t, dssIntent.Reference.SubscriptionId, savedIntent.SubscriptionId)
	require.EqualValues(t, "x", dss.Subscriptions[savedIntent.SubscriptionId].UssBaseUrl)

	_, err := uuid.Parse(string(savedIntent.Ovn))
	require.NoError(t, err)

	timeStart, _ := time.Parse(time.RFC3339Nano, dssIntent.Reference.TimeStart.Value)
	timeEnd, _ := time.Parse(time.RFC3339Nano, dssIntent.Reference.TimeEnd.Value)
	require.Equal(t, timeStart, savedIntent.TimeStart)
	require.Equal(t, timeEnd, savedIntent.TimeEnd)

	require.Equal(t, *dssIntent.Details.Volumes, savedIntent.Volumes)

	require.Len(t, memoryDb.Flights, 1)
	require.Contains(t, memoryDb.Flights, flightId)
	require.Equal(t, dssIntent.Reference.Id, memoryDb.Flights[flightId].EntityID)
}

func TestUpdateFlightPlanSucceeds(t *testing.T) {
	flightId, flight := newFlightParams()
	createResponse := httptest.NewRecorder()
	handler, dss := newHandler()
	memoryDb := handler.DB.(*db.InMemoryDB)
	handler.PutFlightPlan(createResponse, putFlightPlanRequest(&flightId, flight))

	testutil.RequireJSON(t, createResponse, map[string]any{
		"planning_result":    "Completed",
		"flight_plan_status": "Planned",
	})

	flight1 := memoryDb.GetFlight(flightId)
	intent1 := memoryDb.GetIntent(flight1.EntityID)

	updateResponse := httptest.NewRecorder()
	flight.FlightPlan.BasicInformation.Area[0].Volume.AltitudeLower.Value += 1
	handler.PutFlightPlan(updateResponse, putFlightPlanRequest(&flightId, flight))

	testutil.RequireJSON(t, updateResponse, map[string]any{
		"planning_result":    "Completed",
		"flight_plan_status": "OkToFly",
	})

	require.Len(t, dss.Intents, 1)
	require.Len(t, memoryDb.Flights, 1)
	require.Len(t, memoryDb.Intents, 1)

	flight2 := memoryDb.GetFlight(flightId)
	intent2 := memoryDb.GetIntent(flight2.EntityID)

	require.Equal(t, flight1, flight2)
	require.NotZero(t, intent1.Ovn)
	require.Equal(t, intent1.Ovn, intent2.Ovn)

	dssIntent := slices.Collect(maps.Values(dss.Intents))[0]
	require.Equal(t, flight2.EntityID, dssIntent.Reference.Id)
	require.Equal(t, flight.FlightPlan.BasicInformation.Area, *dssIntent.Details.Volumes)
}

func TestPutFlightPlanTooFarOut(t *testing.T) {
	flightId, flight := newFlightParams()
	response := httptest.NewRecorder()
	tooLate := time.Now().Add(time.Hour * 24 * 30).Add(time.Second)
	area := flight.FlightPlan.BasicInformation.Area[0]
	area.TimeStart.Value = tooLate.Format(time.RFC3339Nano)
	area.TimeEnd.Value = tooLate.Add(time.Hour).Format(time.RFC3339Nano)
	planner, _ := newHandler()
	planner.PutFlightPlan(response, putFlightPlanRequest(&flightId, flight))
	testutil.RequireJSON(t, response, map[string]any{
		"activity_result":    "Rejected",
		"planning_result":    "Rejected",
		"flight_plan_status": "NotPlanned",
	})
}

func TestPutAlreadyEndedFlightPlan(t *testing.T) {
	flightId, flight := newFlightParams()
	response := httptest.NewRecorder()
	area := flight.FlightPlan.BasicInformation.Area[0]
	oneSecondAgo := time.Now().Add(-time.Second)
	area.TimeStart.Value = oneSecondAgo.Add(-time.Second).Format(time.RFC3339)
	area.TimeEnd.Value = oneSecondAgo.Format(time.RFC3339)
	planner, _ := newHandler()
	planner.PutFlightPlan(response, putFlightPlanRequest(&flightId, flight))
	testutil.RequireJSON(t, response, map[string]any{
		"activity_result":    "Rejected",
		"planning_result":    "Rejected",
		"flight_plan_status": "NotPlanned",
	})
}

func TestPutRejectsWhenAnotherIntentExists(t *testing.T) {
	flightId, flight := newFlightParams()
	response := httptest.NewRecorder()
	planner, _ := newHandler()
	planner.DB.SaveIntent(db.OperationalIntent{
		EntityID: scdussv1.EntityID(uuid.New().String()),
	})
	planner.PutFlightPlan(response, putFlightPlanRequest(&flightId, flight))
	testutil.RequireJSON(t, response, map[string]any{
		"activity_result":    "Rejected",
		"planning_result":    "Rejected",
		"flight_plan_status": "NotPlanned",
	})
}

func TestUsageStateInUseActivatesFlightPlan(t *testing.T) {
	response := httptest.NewRecorder()
	flightId, flight := newFlightParams()
	flight.FlightPlan.BasicInformation.UsageState = "InUse"
	request := putFlightPlanRequest(&flightId, flight)
	handler, dss := newHandler()
	handler.PutFlightPlan(response, request)

	testutil.RequireJSON(t, response, map[string]any{
		"planning_result":    "Completed",
		"flight_plan_status": "Planned",
	})

	dssIntent := slices.Collect(maps.Values(dss.Intents))[0]
	require.Equal(t, scdussv1.OperationalIntentState_Activated, dssIntent.Reference.State)
}

func TestCreateIntentRetriesWithPeerOvnsWhenKeyIsMissing(t *testing.T) {
	handler := newHandlerFromPeers(t, []scdussv1.OperationalIntent{
		{
			Reference: scdussv1.OperationalIntentReference{
				Id:         scdussv1.EntityID(uuid.New().String()),
				Ovn:        new(scdussv1.EntityOVN(uuid.New().String())),
				UssBaseUrl: scdussv1.OperationalIntentUssBaseURL("http://uss1.localutm"),
			},
		},
	})

	flightId, flight := newFlightParams()
	response := httptest.NewRecorder()
	request := putFlightPlanRequest(&flightId, flight)
	handler.PutFlightPlan(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Len(t, slices.Collect(handler.DB.GetAllIntents()), 1)
	require.Len(t, slices.Collect(handler.DB.GetAllFlights()), 1)
}

func TestCreateIntentRejectsWithPeerPriority100(t *testing.T) {
	handler := newHandlerFromPeers(t, []scdussv1.OperationalIntent{
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
	})

	flightId, flight := newFlightParams()
	response := httptest.NewRecorder()
	request := putFlightPlanRequest(&flightId, flight)
	handler.PutFlightPlan(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	testutil.RequireJSON(t, response, map[string]any{
		"activity_result":    "Rejected",
		"planning_result":    "Rejected",
		"flight_plan_status": "NotPlanned",
	})
	require.Empty(t, slices.Collect(handler.DB.GetAllIntents()))
	require.Empty(t, slices.Collect(handler.DB.GetAllFlights()))
}

func newHandlerFromPeers(t *testing.T, peers []scdussv1.OperationalIntent) *Handler {
	client := utmclient.New(
		auth.NewInMemoryTokenSource(),
		httptest.NewTestServer(t, dss.NewPeerHandler(peers)).Client(),
	)
	return &Handler{
		DSS: &dss.DSS{
			Host:   "https://dss.localutm",
			Client: client,
		},
		Peer:       peer.New(client),
		DB:         db.NewInMemoryDB(),
		UssBaseUrl: scdussv1.OperationalIntentUssBaseURL("http://openuss.localutm"),
	}
}
