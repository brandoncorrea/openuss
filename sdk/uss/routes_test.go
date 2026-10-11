package uss_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/sdk/api"
	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/internal/wiretest"
	"bwawan.com/openuss/sdk/scd"
	"bwawan.com/openuss/sdk/uss"
	"github.com/stretchr/testify/require"
)

type protectedRoute struct {
	method   string
	path     string
	pattern  string
	handler  http.HandlerFunc
	security []api.AuthorizationOption
}

func TestRoutes(t *testing.T) {
	authorizer := &stubAuthorizer{}
	handler := uss.New(scd.NewInMemoryIntentStore(), authorizer)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	routes := []protectedRoute{
		{
			method:   http.MethodGet,
			path:     "/uss/v1/operational_intents/FOO_ID",
			pattern:  "GET /uss/v1/operational_intents/{entity_id}",
			handler:  handler.GetOperationalIntent,
			security: scdussv1.GetOperationalIntentDetailsSecurity,
		},
		{
			method:   http.MethodPost,
			path:     "/uss/v1/operational_intents",
			pattern:  "POST /uss/v1/operational_intents",
			handler:  handler.NotifyOperationalIntent,
			security: scdussv1.NotifyOperationalIntentDetailsChangedSecurity,
		},
		{
			method:   http.MethodPost,
			path:     "/uss/v1/reports",
			pattern:  "POST /uss/v1/reports",
			handler:  handler.MakeUSSReport,
			security: scdussv1.MakeUssReportSecurity,
		},
	}
	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			resolved, pattern := mux.Handler(httptest.NewRequest(route.method, route.path, nil))
			require.Equal(t, route.pattern, pattern)
			require.IsType(t, uss.Protected{}, resolved)
			protected := resolved.(uss.Protected)
			wiretest.RequireSameFunc(t, route.handler, protected.Next)
			require.Equal(t, route.security, protected.Options)
			require.Same(t, authorizer, protected.Authorizer)
		})
	}
}
