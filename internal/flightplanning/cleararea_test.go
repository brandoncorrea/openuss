package flightplanning

import (
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/internal/testutil"
)

func TestClearAreaRequest(t *testing.T) {
	response := httptest.NewRecorder()
	director := &Handler{}
	director.ClearAreaRequests(response, nil)
	testutil.RequireJSON(t, response, map[string]any{
		"outcome": map[string]any{
			"success": true,
		},
	})
}
