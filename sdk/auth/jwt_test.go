package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bwawan.com/openuss/sdk/api"
	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/auth"
	"bwawan.com/openuss/sdk/internal/util"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func newPrivateKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

func newPublicKey(t *testing.T) *rsa.PublicKey {
	t.Helper()
	return &newPrivateKey(t).PublicKey
}

func withClaims(t *testing.T, key *rsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
	require.NoError(t, err)
	return token
}

func newAuthorizationRequirement(scopes ...api.RequiredScope) api.AuthorizationOption {
	return api.AuthorizationOption{
		"Authority": scopes,
	}
}

func newAuthorizationOptions(options [][]api.RequiredScope) []api.AuthorizationOption {
	return util.Map(options, func(scopes []api.RequiredScope) api.AuthorizationOption {
		return newAuthorizationRequirement(scopes...)
	})
}

func signToken(t *testing.T, method jwt.SigningMethod, key *rsa.PrivateKey) string {
	t.Helper()
	token, err := jwt.NewWithClaims(method, jwt.MapClaims{}).SignedString(key)
	require.NoError(t, err)
	return token
}

func withScope(t *testing.T, key *rsa.PrivateKey, requiredScopes ...api.RequiredScope) string {
	t.Helper()
	scopes := util.Map(requiredScopes, util.TrimString)
	claims := jwt.MapClaims{"scope": strings.Join(scopes, " ")}
	return withClaims(t, key, claims)
}

func emptyRequest() *http.Request {
	return httptest.NewRequest(http.MethodGet, "/", nil)
}

func newRequest(token string) *http.Request {
	request := emptyRequest()
	request.Header.Set("Authorization", "Bearer "+token)
	return request
}

func authorize(authorizer api.Authorizer, r *http.Request, options []api.AuthorizationOption) api.AuthorizationResult {
	return authorizer.Authorize(httptest.NewRecorder(), r, options)
}

func authorizeGetDetails(authorizer api.Authorizer, r *http.Request) api.AuthorizationResult {
	options := newAuthorizationOptions([][]api.RequiredScope{
		{scdussv1.UtmStrategicCoordinationScope},
	})
	return authorize(authorizer, r, options)
}

func TestRejectsWithoutBearerToken(t *testing.T) {
	authorizer := auth.NewJWTAuthorizer(newPublicKey(t))
	request := emptyRequest()
	result := authorizeGetDetails(authorizer, request)
	require.ErrorIs(t, result.Error, auth.ErrUnauthenticated)
}

func TestAuthorizesTokenSignedByTheAuthority(t *testing.T) {
	key := newPrivateKey(t)
	authorizer := auth.NewJWTAuthorizer(&key.PublicKey)
	token := withScope(t, key, scdussv1.UtmStrategicCoordinationScope)
	request := newRequest(token)
	result := authorizeGetDetails(authorizer, request)
	require.NoError(t, result.Error)
}

func TestAuthorizesTokenHoldingTheRequiredScopeAmongOthers(t *testing.T) {
	key := newPrivateKey(t)
	authorizer := auth.NewJWTAuthorizer(&key.PublicKey)
	token := withScope(t, key, scdussv1.UtmAvailabilityArbitrationScope, scdussv1.UtmStrategicCoordinationScope)
	request := newRequest(token)
	result := authorizeGetDetails(authorizer, request)
	require.NoError(t, result.Error)
}

func TestAuthorizesTokenSatisfyingAnyOneOption(t *testing.T) {
	key := newPrivateKey(t)
	authorizer := auth.NewJWTAuthorizer(&key.PublicKey)
	token := withScope(t, key, scdussv1.UtmConstraintManagementScope)
	request := newRequest(token)
	options := newAuthorizationOptions([][]api.RequiredScope{
		{scdussv1.UtmStrategicCoordinationScope},
		{scdussv1.UtmConstraintManagementScope},
		{scdussv1.UtmAvailabilityArbitrationScope},
	})
	result := authorize(authorizer, request, options)
	require.NoError(t, result.Error)
}

func TestRejectsTokenSignedByAnotherKey(t *testing.T) {
	authorizer := auth.NewJWTAuthorizer(newPublicKey(t))
	token := signToken(t, jwt.SigningMethodRS256, newPrivateKey(t))
	request := newRequest(token)
	result := authorizeGetDetails(authorizer, request)
	require.ErrorIs(t, result.Error, auth.ErrUnauthenticated)
}

func TestRejectsValidTokenMissingBearerPrefix(t *testing.T) {
	key := newPrivateKey(t)
	authorizer := auth.NewJWTAuthorizer(&key.PublicKey)
	token := signToken(t, jwt.SigningMethodRS256, key)
	request := emptyRequest()
	request.Header.Set("Authorization", token)
	result := authorizeGetDetails(authorizer, request)
	require.ErrorIs(t, result.Error, auth.ErrUnauthenticated)
}

func TestRejectsTokenSignedWithAnotherAlgorithm(t *testing.T) {
	key := newPrivateKey(t)
	authorizer := auth.NewJWTAuthorizer(&key.PublicKey)
	token := signToken(t, jwt.SigningMethodRS512, key)
	request := newRequest(token)
	result := authorizeGetDetails(authorizer, request)

	require.ErrorIs(t, result.Error, auth.ErrUnauthenticated)
}

func TestRejectsTokenWithoutTheRequiredScope(t *testing.T) {
	key := newPrivateKey(t)
	authorizer := auth.NewJWTAuthorizer(&key.PublicKey)
	token := withScope(t, key, scdussv1.UtmStrategicCoordinationScope)
	request := newRequest(token)
	options := newAuthorizationOptions([][]api.RequiredScope{
		{scdussv1.UtmConstraintManagementScope},
	})
	result := authorize(authorizer, request, options)
	require.ErrorIs(t, result.Error, auth.ErrPermissionDenied)
}

func TestRejectsTokenOnlyOneOfManyRequiredScopes(t *testing.T) {
	key := newPrivateKey(t)
	authorizer := auth.NewJWTAuthorizer(&key.PublicKey)
	token := withScope(t, key, scdussv1.UtmStrategicCoordinationScope)
	request := newRequest(token)
	options := newAuthorizationOptions([][]api.RequiredScope{
		{
			scdussv1.UtmStrategicCoordinationScope,
			scdussv1.UtmConstraintManagementScope,
		},
	})
	result := authorize(authorizer, request, options)
	require.ErrorIs(t, result.Error, auth.ErrPermissionDenied)
}
