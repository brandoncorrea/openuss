package flightplanning

import (
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/internal/testutil"
)

func TestGetStatus(t *testing.T) {
	response := httptest.NewRecorder()
	director := &Handler{}
	director.GetStatus(response, nil)
	testutil.RequireJSON(t, response, map[string]any{
		"status": "Ready",
	})
}
