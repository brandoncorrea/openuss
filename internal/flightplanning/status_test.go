package flightplanning_test

import (
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/internal/flightplanning"
	"bwawan.com/openuss/internal/wiretest"
)

func TestGetStatus(t *testing.T) {
	response := httptest.NewRecorder()
	director := &flightplanning.Handler{}
	director.GetStatus(response, nil)
	wiretest.RequireJSON(t, response, map[string]any{
		"status": "Ready",
	})
}
