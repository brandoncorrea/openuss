package flightplanning

import (
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/internal/testutil"
)

func TestPutFlightPlan(t *testing.T) {
	response := httptest.NewRecorder()
	PutFlightPlan(response, nil)
	testutil.RequireJSON(t, response, map[string]any{
		"planning_result":    "Rejected",
		"flight_plan_status": "NotPlanned",
	})
}
