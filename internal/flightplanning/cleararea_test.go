package flightplanning_test

import (
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/internal/flightplanning"
	"bwawan.com/openuss/internal/wiretest"
)

func TestClearAreaRequest(t *testing.T) {
	response := httptest.NewRecorder()
	director := &flightplanning.Handler{}
	director.ClearAreaRequests(response, nil)
	wiretest.RequireJSON(t, response, map[string]any{
		"outcome": map[string]any{
			"success": true,
		},
	})
}
