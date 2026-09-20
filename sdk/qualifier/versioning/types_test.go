package versioning_test

import (
	"testing"

	"bwawan.com/openuss/sdk/internal/wiretest"
	"bwawan.com/openuss/sdk/qualifier/versioning"
)

func TestGetVersionResponse(t *testing.T) {
	response := versioning.GetVersionResponse{
		SystemIdentity: "the-identity",
		SystemVersion:  "the-version",
	}
	json := map[string]any{
		"system_identity": "the-identity",
		"system_version":  "the-version",
	}
	wiretest.RequireJSONEncoding(t, json, response)
}
