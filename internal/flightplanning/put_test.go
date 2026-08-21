package flightplanning

import (
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/dss"
	"bwawan.com/openuss/internal/testutil"
	"github.com/stretchr/testify/require"
)

func newFlightPlanBody() PutFlightPlanBody {
	return PutFlightPlanBody{
		RequestId: "e4d3a1b2-5c6d-4e7f-8a9b-0c1d2e3f4a5b",
		FlightPlan: FlightPlan{
			BasicInformation: FlightPlanBasicInformation{
				Area: []scdussv1.Volume4D{
					{
						TimeStart: &scdussv1.Time{
							Value:  "2026-08-13T00:00:00Z",
							Format: "RFC3339",
						},
						TimeEnd: &scdussv1.Time{
							Value:  "2026-08-13T01:00:00Z",
							Format: "RFC3339",
						},
						Volume: scdussv1.Volume3D{
							OutlineCircle: &scdussv1.Circle{
								Radius: &scdussv1.Radius{
									Value: 100,
									Units: "M",
								},
								Center: &scdussv1.LatLngPoint{
									Lng: -80.6,
									Lat: 37.2,
								},
							},
							AltitudeLower: &scdussv1.Altitude{
								Value:     0,
								Reference: "W84",
								Units:     "M",
							},
							AltitudeUpper: &scdussv1.Altitude{
								Value:     100,
								Reference: "W84",
								Units:     "M",
							},
						},
					},
				},
			},
		},
	}
}

func putFlightPlanRequest(body any) *http.Request {
	bytes, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	return httptest.NewRequest(
		http.MethodPut,
		"/flight_planning/v1/flight_plans/e4d3a1b2-5c6d-4e7f-8a9b-0c1d2e3f4a5b",
		strings.NewReader(string(bytes)),
	)
}

func newHandler() (*Handler, *dss.InMemoryDSS) {
	dss := dss.NewInMemoryDSS()
	return &Handler{DSS: dss}, dss
}

func TestPutFlightPlanSucceeds(t *testing.T) {
	response := httptest.NewRecorder()
	flight := newFlightPlanBody()
	area := flight.FlightPlan.BasicInformation.Area[0]
	area.TimeStart.Value = time.Now().Add(time.Hour).Format(time.RFC3339)
	area.TimeEnd.Value = time.Now().Add(2 * time.Hour).Format(time.RFC3339)
	planner, dss := newHandler()
	planner.PutFlightPlan(response, putFlightPlanRequest(flight))

	require.Len(t, dss.References, 1)
	for id, reference := range dss.References {
		require.NotZero(t, id)
		require.Equal(t, flight.FlightPlan.BasicInformation.Area, reference.Extents)
		require.Equal(t, scdussv1.OperationalIntentState_Accepted, reference.State)
		require.Equal(t, "http://host.docker.internal:8080", string(reference.UssBaseUrl))
	}
	// stored := dss.References[]
	// require.Equal()

	testutil.RequireJSON(t, response, map[string]any{
		"planning_result":    "Completed",
		"flight_plan_status": "Planned",
	})
}

func TestPutFlightPlanTooFarOut(t *testing.T) {
	response := httptest.NewRecorder()
	flight := newFlightPlanBody()
	tooLate := time.Now().Add(time.Hour * 24 * 30).Add(time.Second)
	area := flight.FlightPlan.BasicInformation.Area[0]
	area.TimeStart.Value = tooLate.Format(time.RFC3339)
	area.TimeEnd.Value = tooLate.Add(time.Hour).Format(time.RFC3339)
	planner, _ := newHandler()
	planner.PutFlightPlan(response, putFlightPlanRequest(flight))
	testutil.RequireJSON(t, response, map[string]any{
		"activity_result":    "Rejected",
		"planning_result":    "Rejected",
		"flight_plan_status": "NotPlanned",
	})
}

func TestPutAlreadyEndedFlightPlan(t *testing.T) {
	response := httptest.NewRecorder()
	flight := newFlightPlanBody()
	area := flight.FlightPlan.BasicInformation.Area[0]
	oneSecondAgo := time.Now().Add(-time.Second)
	area.TimeStart.Value = oneSecondAgo.Add(-time.Second).Format(time.RFC3339)
	area.TimeEnd.Value = oneSecondAgo.Format(time.RFC3339)
	planner, _ := newHandler()
	planner.PutFlightPlan(response, putFlightPlanRequest(flight))
	testutil.RequireJSON(t, response, map[string]any{
		"activity_result":    "Rejected",
		"planning_result":    "Rejected",
		"flight_plan_status": "NotPlanned",
	})
}
