package flightplanning

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/testutil"
	"github.com/stretchr/testify/require"
)

func newDeleteRequest(flightPlanId *string) *http.Request {
	request := httptest.NewRequest(http.MethodDelete, "/blah", nil)
	if flightPlanId != nil {
		request.SetPathValue("flight_plan_id", *flightPlanId)
	}
	return request
}

func newIntent() scdussv1.OperationalIntent {
	return scdussv1.OperationalIntent{
		Reference: scdussv1.OperationalIntentReference{
			Id: scdussv1.EntityID(uuid.New().String()),
		},
	}
}

func TestDeleteFlightPlanSucceeds(t *testing.T) {
	handler, _ := newHandler()
	intent := newIntent()
	handler.DB.SaveIntent(intent)

	recorder := httptest.NewRecorder()
	request := newDeleteRequest(new(string(intent.Reference.Id)))
	handler.DeleteFlightPlan(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Nil(t, handler.DB.GetIntent(intent.Reference.Id))
	testutil.RequireJSON(t, recorder, map[string]any{
		"flight_plan_status": "Closed",
		"planning_result":    "Completed",
	})
}

func TestDeleteFlightPlanMissingFlightId(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := newDeleteRequest(nil)
	handler, _ := newHandler()
	handler.DeleteFlightPlan(recorder, request)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDeleteFlightPlanDoesNotExist(t *testing.T) {
	recorder := httptest.NewRecorder()
	id := new(uuid.New().String())
	request := newDeleteRequest(id)
	request.SetPathValue("flight_plan_id", uuid.New().String())
	handler, _ := newHandler()
	handler.DeleteFlightPlan(recorder, request)
	require.Equal(t, http.StatusNotFound, recorder.Code)
}
