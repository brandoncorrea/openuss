package uss_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/sdk/api"
	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/auth"
	"bwawan.com/openuss/sdk/internal/wiretest"
	"bwawan.com/openuss/sdk/uss"
	"github.com/stretchr/testify/require"
)

type stubAuthorizer struct {
	err      error
	response http.ResponseWriter
	request  *http.Request
	options  []api.AuthorizationOption
}

func (s *stubAuthorizer) Authorize(w http.ResponseWriter, r *http.Request, options []api.AuthorizationOption) api.AuthorizationResult {
	s.response = w
	s.request = r
	s.options = options
	return api.AuthorizationResult{Error: s.err}
}

func noContent(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func TestAuthorizesTheRequestAgainstTheRouteOptions(t *testing.T) {
	authorizer := &stubAuthorizer{}
	protected := uss.Protected{
		Authorizer: authorizer,
		Options:    scdussv1.MakeUssReportSecurity,
		Next:       noContent,
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	protected.ServeHTTP(response, request)
	require.Same(t, response, authorizer.response)
	require.Same(t, request, authorizer.request)
	require.Equal(t, scdussv1.MakeUssReportSecurity, authorizer.options)
}

func TestPassesAnAuthorizedRequestToTheHandler(t *testing.T) {
	protected := uss.Protected{
		Authorizer: auth.NewUnsafeAuthorizer(),
		Next:       noContent,
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	protected.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestRejectsAnUnauthenticatedRequestWith401(t *testing.T) {
	protected := uss.Protected{
		Authorizer: &stubAuthorizer{err: auth.ErrUnauthenticated},
		Next:       wiretest.AssertNotCalledHandler(t),
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	protected.ServeHTTP(response, request)
	wiretest.RequireJSONStatus(t, response, http.StatusUnauthorized, scdussv1.ErrorResponse{})
}

func TestRejectsARequestWithoutPermissionWith403(t *testing.T) {
	protected := uss.Protected{
		Authorizer: &stubAuthorizer{err: auth.ErrPermissionDenied},
		Next:       wiretest.AssertNotCalledHandler(t),
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	protected.ServeHTTP(response, request)
	wiretest.RequireJSONStatus(t, response, http.StatusForbidden, scdussv1.ErrorResponse{})
}
