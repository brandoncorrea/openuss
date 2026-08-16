package flightplanning

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestDeleteFlightPlan(t *testing.T) {
	recorder := httptest.NewRecorder()
	handler := &Handler{}
	handler.DeleteFlightPlan(recorder, nil)
	require.Equal(t, http.StatusOK, recorder.Code)
	testutil.RequireJSON(t, recorder, map[string]any{
		"flight_plan_status": "Closed",
		"planning_result":    "Completed",
	})
}
