package auth

import (
	"net/http"

	"bwawan.com/openuss/sdk/api"
)

type UnsafeAuthorizer struct{}

func NewUnsafeAuthorizer() *UnsafeAuthorizer {
	return &UnsafeAuthorizer{}
}

func (*UnsafeAuthorizer) Authorize(http.ResponseWriter, *http.Request, []api.AuthorizationOption) api.AuthorizationResult {
	return api.AuthorizationResult{}
}
