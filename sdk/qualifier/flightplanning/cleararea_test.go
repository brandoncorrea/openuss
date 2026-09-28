package flightplanning_test

import (
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/sdk/internal/wiretest"
	"bwawan.com/openuss/sdk/qualifier/flightplanning"
)

func TestClearAreaRequestsReportSuccess(t *testing.T) {
	response := httptest.NewRecorder()
	handler := &flightplanning.Handler{}
	handler.ClearAreaRequests(response, nil)
	wiretest.RequireJSON(t, response, flightplanning.ClearAreaResponse{
		Outcome: flightplanning.ClearAreaOutcome{
			Success: true,
		},
	})
}
