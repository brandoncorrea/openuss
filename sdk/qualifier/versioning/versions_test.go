package versioning_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/sdk/internal/wiretest"
	"bwawan.com/openuss/sdk/qualifier/versioning"
)

func TestGetVersionReportsSystemIdentityAndVersion(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.SetPathValue("system_identity", "foo")
	versioning.New().GetVersion(response, request)
	wiretest.RequireJSON(t, response, versioning.GetVersionResponse{
		SystemIdentity: "foo",
		SystemVersion:  "blah",
	})
}
