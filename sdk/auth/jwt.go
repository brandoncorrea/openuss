package auth

import (
	"crypto/rsa"
	"encoding/json/v2"
	"errors"
	"net/http"
	"slices"
	"strings"

	"bwawan.com/openuss/sdk/api"
	"bwawan.com/openuss/sdk/internal/util"
	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthorizer struct {
	PublicKey *rsa.PublicKey
}

type GrantedScopes []api.RequiredScope

// TODO(gap): If a claim has two "scope" keys, this gets called twice,
// producing the union of the two scopes.
func (s *GrantedScopes) UnmarshalJSON(data []byte) error {
	var spaceSeparated string
	// TODO(gap): what if this fails to unmarshal?
	json.Unmarshal(data, &spaceSeparated)
	for scope := range strings.FieldsSeq(spaceSeparated) {
		*s = append(*s, api.RequiredScope(scope))
	}
	return nil
}

type JWTClaims struct {
	jwt.RegisteredClaims
	Scopes GrantedScopes `json:"scope"`
}

func NewJWTAuthorizer(key *rsa.PublicKey) *JWTAuthorizer {
	return &JWTAuthorizer{PublicKey: key}
}

var (
	ErrUnauthenticated  = errors.New("auth: unauthenticated")
	ErrPermissionDenied = errors.New("auth: permission denied")
	ErrMissingToken     = errors.New("auth: token not supplied")
	ErrWrongScope       = errors.New("auth: wrong scope supplied")
)

// TODO(gap): does not check aud, iss, sub, or exp claims
func (a *JWTAuthorizer) Authorize(_ http.ResponseWriter, r *http.Request, options []api.AuthorizationOption) api.AuthorizationResult {
	token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !found {
		return api.AuthorizationResult{Error: errors.Join(ErrUnauthenticated, ErrMissingToken)}
	}
	claims, err := a.parseToken(token)
	if err != nil {
		return api.AuthorizationResult{Error: errors.Join(ErrUnauthenticated, err)}
	}
	if !hasAuthorizedScope(claims.Scopes, options) {
		return api.AuthorizationResult{
			Error: errors.Join(ErrPermissionDenied, ErrWrongScope),
		}
	}
	return api.AuthorizationResult{}
}

func hasAuthorizedScope(granted []api.RequiredScope, options []api.AuthorizationOption) bool {
	return slices.ContainsFunc(options, func(option api.AuthorizationOption) bool {
		// TODO: api.AuthorizationOption may not always contain "Authority"
		// if Authority is missing, then access is granted.
		return util.EveryFunc(option["Authority"], func(required api.RequiredScope) bool {
			return slices.Contains(granted, required)
		})
	})
}

func (a *JWTAuthorizer) parseToken(token string) (*JWTClaims, error) {
	var claims JWTClaims
	_, err := jwt.ParseWithClaims(token, &claims, a.publicKey, jwt.WithValidMethods([]string{
		jwt.SigningMethodRS256.Alg(),
	}))
	if err != nil {
		return nil, err
	}
	return &claims, nil
}

func (a *JWTAuthorizer) publicKey(*jwt.Token) (any, error) {
	return a.PublicKey, nil
}
