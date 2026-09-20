package wiretest

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

type Registrar interface {
	RegisterRoutes(*http.ServeMux)
}

type RouteRegistration struct {
	Method  string
	Path    string
	Pattern string
	Handler http.HandlerFunc
}

func RequireRouteRegistration(
	t *testing.T,
	registrar Registrar,
	registrations []RouteRegistration,
) {
	t.Helper()
	mux := http.NewServeMux()
	registrar.RegisterRoutes(mux)

	for _, route := range registrations {
		t.Run(route.Method+" "+route.Path, func(t *testing.T) {
			resolved, pattern := mux.Handler(httptest.NewRequest(route.Method, route.Path, nil))
			require.Equal(t, route.Pattern, pattern)
			requireSameFunc(t, route.Handler, resolved)
		})
	}
}

func requireSameFunc(t *testing.T, want, got any) {
	t.Helper()
	wantValue := resolvePointer(want)
	gotValue := resolvePointer(got)
	failureMessage := "Expected: " + resolveName(wantValue) + ", got: " + resolveName(gotValue)
	require.Equal(t, wantValue, gotValue, failureMessage)
}

func resolveName(ptr uintptr) string {
	return runtime.FuncForPC(ptr).Name()
}

func resolvePointer(f any) uintptr {
	return reflect.ValueOf(f).Pointer()
}
