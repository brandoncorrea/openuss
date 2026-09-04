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

func newOperationalIntent(t *testing.T, handler *Handler) scdussv1.OperationalIntentReference {
	id := scdussv1.EntityID(uuid.New().String())
	params := scdussv1.PutOperationalIntentReferenceParameters{}
	result, _ := handler.DSS.CreateOperationalIntentReference(t.Context(), id, params)
	reference := result.OperationalIntentReference
	handler.DB.SaveIntent(scdussv1.OperationalIntent{
		Reference: reference,
	})
	return reference
}

func TestDeleteFlightPlanSucceeds(t *testing.T) {
	handler, dss := newHandler()
	reference := newOperationalIntent(t, handler)

	recorder := httptest.NewRecorder()
	request := newDeleteRequest(new(string(reference.Id)))
	handler.DeleteFlightPlan(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Nil(t, handler.DB.GetIntent(reference.Id))
	require.NotContains(t, dss.Intents, reference.Id)
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

func TestDeleteFlightPlanFails(t *testing.T) {
	handler, dss := newHandler()
	reference := newOperationalIntent(t, handler)
	intent := handler.DB.GetIntent(reference.Id)
	intent.Reference.Ovn = new(scdussv1.EntityOVN(uuid.New().String()))
	handler.DB.SaveIntent(*intent)

	recorder := httptest.NewRecorder()
	request := newDeleteRequest(new(string(reference.Id)))
	handler.DeleteFlightPlan(recorder, request)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Equal(t, intent, handler.DB.GetIntent(reference.Id))
	require.Contains(t, dss.Intents, reference.Id)
}
