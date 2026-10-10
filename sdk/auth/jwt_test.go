package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/sdk/api"
	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/auth"
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

func signToken(t *testing.T, method jwt.SigningMethod, key *rsa.PrivateKey) string {
	t.Helper()
	token, err := jwt.NewWithClaims(method, jwt.MapClaims{}).SignedString(key)
	require.NoError(t, err)
	return token
}

func emptyRequest() *http.Request {
	return httptest.NewRequest(http.MethodGet, "/", nil)
}

func newRequest(token string) *http.Request {
	request := emptyRequest()
	request.Header.Set("Authorization", "Bearer "+token)
	return request
}

func authorizeGetDetails(authorizer api.Authorizer, r *http.Request) api.AuthorizationResult {
	response := httptest.NewRecorder()
	return authorizer.Authorize(response, r, scdussv1.GetOperationalIntentDetailsSecurity)
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
	token := signToken(t, jwt.SigningMethodRS256, key)
	request := newRequest(token)
	result := authorizeGetDetails(authorizer, request)
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
