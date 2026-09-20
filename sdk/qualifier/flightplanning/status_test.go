package flightplanning_test

import (
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/sdk/internal/wiretest"
	"bwawan.com/openuss/sdk/qualifier/flightplanning"
)

func TestGetStatus(t *testing.T) {
	response := httptest.NewRecorder()
	director := &flightplanning.Handler{}
	director.GetStatus(response, nil)
	wiretest.RequireJSON(t, response, map[string]any{
		"status": "Ready",
	})
}
