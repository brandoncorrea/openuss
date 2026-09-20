package versioning_test

import (
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/sdk/internal/wiretest"
	"bwawan.com/openuss/sdk/qualifier/versioning"
)

func TestGetVersion(t *testing.T) {
	response := httptest.NewRecorder()
	versioning.New().GetVersion(response, nil)
	wiretest.RequireJSON(t, response, versioning.GetVersionResponse{
		SystemIdentity: "astm.f3548.v21",
		SystemVersion:  "blah",
	})
}
