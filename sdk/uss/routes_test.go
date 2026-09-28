package uss_test

import (
	"net/http"
	"testing"

	"bwawan.com/openuss/sdk/internal/wiretest"
	"bwawan.com/openuss/sdk/scd"
	"bwawan.com/openuss/sdk/uss"
)

func TestRoutes(t *testing.T) {
	handler := uss.New(scd.NewInMemoryIntentStore())
	wiretest.RequireRouteRegistration(t, handler, []wiretest.RouteRegistration{
		{
			Method:  http.MethodGet,
			Path:    "/uss/v1/operational_intents/FOO_ID",
			Pattern: "GET /uss/v1/operational_intents/{entity_id}",
			Handler: handler.GetOperationalIntent,
		},
		{
			Method:  http.MethodPost,
			Path:    "/uss/v1/operational_intents",
			Pattern: "POST /uss/v1/operational_intents",
			Handler: handler.NotifyOperationalIntent,
		},
		{
			Method:  http.MethodPost,
			Path:    "/uss/v1/reports",
			Pattern: "POST /uss/v1/reports",
			Handler: handler.MakeUSSReport,
		},
	})
}
