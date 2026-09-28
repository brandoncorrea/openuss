package flightplanning_test

import (
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/sdk/internal/wiretest"
	"bwawan.com/openuss/sdk/qualifier/flightplanning"
)

func TestGetStatusReportsReady(t *testing.T) {
	response := httptest.NewRecorder()
	handler := &flightplanning.Handler{}
	handler.GetStatus(response, nil)
	wiretest.RequireJSON(t, response, flightplanning.StatusResponse{
		Status: flightplanning.ServiceStatusReady,
	})
}
