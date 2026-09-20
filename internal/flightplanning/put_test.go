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

	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/auth"
	"bwawan.com/openuss/internal/db"
	"bwawan.com/openuss/internal/dss"
	"bwawan.com/openuss/internal/dss/dsstest"
	"bwawan.com/openuss/internal/flightplanning"
	"bwawan.com/openuss/internal/peer"
	"bwawan.com/openuss/internal/scd"
	"bwawan.com/openuss/internal/scdtest"
	"bwawan.com/openuss/internal/utmclient"
	"bwawan.com/openuss/internal/wiretest"
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
	authority := dss.NewInMemoryDSS()
	handler := flightplanning.New(
		authority,
		nil,
		db.NewInMemoryDB(),
		scd.NewInMemoryIntentStore(),
		"http://openuss.localutm",
	)
	return handler, authority
}

func TestCreateFlightPlanSucceeds(t *testing.T) {
	flightID, flight := newFlightParams()
	response := httptest.NewRecorder()
	request := putFlightPlanRequest(&flightID, flight)
	handler, authority := newHandler()
	handler.PutFlightPlan(response, request)

	wiretest.RequireJSON(t, response, map[string]any{
		"planning_result":    "Completed",
		"flight_plan_status": "Planned",
	})

	require.Len(t, authority.Intents, 1)

	dssIntent := slices.Collect(maps.Values(authority.Intents))[0]
	require.Equal(t, flight.FlightPlan.BasicInformation.Area, *dssIntent.Details.Volumes)
	require.Equal(t, scdussv1.OperationalIntentState_Accepted, dssIntent.Reference.State)
	require.EqualValues(t, "http://openuss.localutm", dssIntent.Reference.UssBaseUrl)

	savedIntent, err := handler.Intents.Get(t.Context(), dssIntent.Reference.Id)
	require.NoError(t, err)
	require.Equal(t, dssIntent.Reference.Id, savedIntent.EntityID)
	require.Equal(t, "InMemoryManager", savedIntent.Manager)
	require.Equal(t, scdussv1.UssAvailabilityState_Normal, savedIntent.USSAvailability)
	require.EqualValues(t, 1, savedIntent.Version)
	require.EqualValues(t, 2, savedIntent.Priority)
	require.Equal(t, scdussv1.OperationalIntentState_Accepted, savedIntent.State)
	require.Equal(t, dssIntent.Reference.SubscriptionId, savedIntent.SubscriptionID)
	require.EqualValues(t, "x", authority.Subscriptions[savedIntent.SubscriptionID].UssBaseUrl)

	_, err = uuid.Parse(string(savedIntent.OVN))
	require.NoError(t, err)

	timeStart, _ := time.Parse(time.RFC3339Nano, dssIntent.Reference.TimeStart.Value)
	timeEnd, _ := time.Parse(time.RFC3339Nano, dssIntent.Reference.TimeEnd.Value)
	require.Equal(t, timeStart, savedIntent.TimeStart)
	require.Equal(t, timeEnd, savedIntent.TimeEnd)

	require.Equal(t, *dssIntent.Details.Volumes, savedIntent.Volumes)

	require.Len(t, slices.Collect(handler.DB.GetAllFlights()), 1)
	require.Equal(t, dssIntent.Reference.Id, handler.DB.GetFlight(flightID).EntityID)
}

func TestUpdateFlightPlanSucceeds(t *testing.T) {
	flightID, flight := newFlightParams()
	createResponse := httptest.NewRecorder()
	handler, authority := newHandler()
	handler.PutFlightPlan(createResponse, putFlightPlanRequest(&flightID, flight))

	wiretest.RequireJSON(t, createResponse, map[string]any{
		"planning_result":    "Completed",
		"flight_plan_status": "Planned",
	})

	flight1 := handler.DB.GetFlight(flightID)
	intent1, err := handler.Intents.Get(t.Context(), flight1.EntityID)
	require.NoError(t, err)

	updateResponse := httptest.NewRecorder()
	flight.FlightPlan.BasicInformation.Area[0].Volume.AltitudeLower.Value += 1
	handler.PutFlightPlan(updateResponse, putFlightPlanRequest(&flightID, flight))

	wiretest.RequireJSON(t, updateResponse, map[string]any{
		"planning_result":    "Completed",
		"flight_plan_status": "OkToFly",
	})

	require.Len(t, authority.Intents, 1)
	require.Len(t, slices.Collect(handler.DB.GetAllFlights()), 1)

	intents, err := handler.Intents.List(t.Context())
	require.NoError(t, err)
	require.Len(t, intents, 1)

	flight2 := handler.DB.GetFlight(flightID)
	intent2, err := handler.Intents.Get(t.Context(), flight2.EntityID)

	require.NoError(t, err)
	require.Equal(t, flight1, flight2)
	require.NotZero(t, intent1.OVN)
	require.Equal(t, intent1.OVN, intent2.OVN)

	dssIntent := slices.Collect(maps.Values(authority.Intents))[0]
	require.Equal(t, flight2.EntityID, dssIntent.Reference.Id)
	require.Equal(t, flight.FlightPlan.BasicInformation.Area, *dssIntent.Details.Volumes)
}

func TestPutFlightPlanTooFarOut(t *testing.T) {
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

func TestPutAlreadyEndedFlightPlan(t *testing.T) {
	flightID, flight := newFlightParams()
	response := httptest.NewRecorder()
	area := flight.FlightPlan.BasicInformation.Area[0]
	oneSecondAgo := time.Now().Add(-time.Second)
	area.TimeStart.Value = oneSecondAgo.Add(-time.Second).Format(time.RFC3339)
	area.TimeEnd.Value = oneSecondAgo.Format(time.RFC3339)
	planner, _ := newHandler()
	planner.PutFlightPlan(response, putFlightPlanRequest(&flightID, flight))
	wiretest.RequireJSON(t, response, map[string]any{
		"activity_result":    "Rejected",
		"planning_result":    "Rejected",
		"flight_plan_status": "NotPlanned",
	})
}

func TestPutRejectsWhenAnotherIntentExists(t *testing.T) {
	flightID, flight := newFlightParams()
	response := httptest.NewRecorder()
	planner, _ := newHandler()
	require.NoError(t, planner.Intents.Upsert(t.Context(), scd.OperationalIntent{
		EntityID: scdussv1.EntityID(uuid.New().String()),
	}))
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
	handler, authority := newHandler()
	handler.PutFlightPlan(response, request)

	wiretest.RequireJSON(t, response, map[string]any{
		"planning_result":    "Completed",
		"flight_plan_status": "Planned",
	})

	dssIntent := slices.Collect(maps.Values(authority.Intents))[0]
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

	flightID, flight := newFlightParams()
	response := httptest.NewRecorder()
	request := putFlightPlanRequest(&flightID, flight)
	handler.PutFlightPlan(response, request)

	require.Equal(t, http.StatusOK, response.Code)

	intents, err := handler.Intents.List(t.Context())
	require.NoError(t, err)
	require.Len(t, intents, 1)
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

	flightID, flight := newFlightParams()
	response := httptest.NewRecorder()
	request := putFlightPlanRequest(&flightID, flight)
	handler.PutFlightPlan(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	wiretest.RequireJSON(t, response, map[string]any{
		"activity_result":    "Rejected",
		"planning_result":    "Rejected",
		"flight_plan_status": "NotPlanned",
	})

	intents, err := handler.Intents.List(t.Context())
	require.NoError(t, err)
	require.Empty(t, intents)
	require.Empty(t, slices.Collect(handler.DB.GetAllFlights()))
}

func TestCreateIntentApprovesWithActivePeerAtPriority100(t *testing.T) {
	handler := newHandlerFromPeers(t, []scdussv1.OperationalIntent{
		{
			Reference: scdussv1.OperationalIntentReference{
				Id:         scdussv1.EntityID(uuid.New().String()),
				Ovn:        new(scdussv1.EntityOVN(uuid.New().String())),
				UssBaseUrl: scdussv1.OperationalIntentUssBaseURL("http://uss1.localutm"),
				State:      scdussv1.OperationalIntentState_Activated,
			},
			Details: scdussv1.OperationalIntentDetails{
				Priority: new(scdussv1.Priority(100)),
			},
		},
	})

	flightID, flight := newFlightParams()
	response := httptest.NewRecorder()
	request := putFlightPlanRequest(&flightID, flight)
	handler.PutFlightPlan(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	wiretest.RequireJSON(t, response, map[string]any{
		"planning_result":    "Completed",
		"flight_plan_status": "Planned",
	})
	plan := handler.DB.GetFlight(flightID)
	require.NotNil(t, plan)

	stored, err := handler.Intents.Get(t.Context(), plan.EntityID)
	require.NoError(t, err)
	require.Equal(t, plan.EntityID, stored.EntityID)
}

func TestUpdateIntentRejectsWithPeerPriority100(t *testing.T) {
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

	flightID, flightParams := newFlightParams()

	intent := scd.OperationalIntent{
		EntityID: scdussv1.EntityID(uuid.New().String()),
		OVN:      scdussv1.EntityOVN(uuid.New().String()),
	}
	flight := db.FlightPlan{
		ID:       flightID,
		EntityID: intent.EntityID,
	}
	require.NoError(t, handler.Intents.Upsert(t.Context(), intent))
	handler.DB.SaveFlight(flight)

	response := httptest.NewRecorder()
	request := putFlightPlanRequest(&flightID, flightParams)
	handler.PutFlightPlan(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	wiretest.RequireJSON(t, response, map[string]any{
		"activity_result":    "Rejected",
		"planning_result":    "Rejected",
		"flight_plan_status": "Planned",
	})

	intents, err := handler.Intents.List(t.Context())
	require.NoError(t, err)
	require.ElementsMatch(t, []scd.OperationalIntent{intent}, intents)
	require.ElementsMatch(t, []db.FlightPlan{flight}, slices.Collect(handler.DB.GetAllFlights()))
}

func newHandlerFromPeers(t *testing.T, peers []scdussv1.OperationalIntent) *flightplanning.Handler {
	client := utmclient.New(
		auth.NewInMemoryTokenSource(),
		httptest.NewTestServer(t, dsstest.NewPeerHandler(peers)).Client(),
	)
	return flightplanning.New(
		dss.New("https://dss.localutm", client),
		peer.New(client),
		db.NewInMemoryDB(),
		scd.NewInMemoryIntentStore(),
		"http://openuss.localutm",
	)
}
