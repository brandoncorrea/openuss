package uss

import (
	"errors"
	"net/http"

	"bwawan.com/openuss/sdk/api"
	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/auth"
)

type Protected struct {
	Authorizer api.Authorizer
	Options    []api.AuthorizationOption
	Next       http.HandlerFunc
}

func (p Protected) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result := p.Authorizer.Authorize(w, r, p.Options)
	if result.Error != nil {
		writeAuthorizationFailure(w, result.Error)
		return
	}
	p.Next(w, r)
}

// TODO(gap): a 401 does not send WWW-Authenticate: Bearer (RFC 6750 section 3)
func writeAuthorizationFailure(w http.ResponseWriter, err error) {
	// TODO(gap): No "Message" supplied to ErrorResponse
	api.WriteJSON(w, authorizationFailureStatus(err), scdussv1.ErrorResponse{})
}

func authorizationFailureStatus(err error) int {
	if errors.Is(err, auth.ErrPermissionDenied) {
		return http.StatusForbidden
	}
	return http.StatusUnauthorized
}
