package versioning

import (
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/internal/testutil"
)

func TestGetVersion(t *testing.T) {
	response := httptest.NewRecorder()
	GetVersion(response, nil)
	testutil.RequireJSON(t, response, map[string]any{
		"system_identity": "astm.f3548.v21",
		"system_version":  "",
	})
}
