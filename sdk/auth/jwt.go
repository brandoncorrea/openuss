package auth

import (
	"crypto/rsa"
	"errors"
	"net/http"
	"strings"

	"bwawan.com/openuss/sdk/api"
	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthorizer struct {
	PublicKey *rsa.PublicKey
}

func NewJWTAuthorizer(key *rsa.PublicKey) *JWTAuthorizer {
	return &JWTAuthorizer{PublicKey: key}
}

var ErrUnauthenticated = errors.New("auth: unauthenticated")

func (a *JWTAuthorizer) Authorize(_ http.ResponseWriter, r *http.Request, _ []api.AuthorizationOption) api.AuthorizationResult {
	token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !found {
		return api.AuthorizationResult{Error: ErrUnauthenticated}
	}
	if _, err := a.parseToken(token); err != nil {
		return api.AuthorizationResult{Error: ErrUnauthenticated}
	}
	return api.AuthorizationResult{}
}

func (a *JWTAuthorizer) parseToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, a.publicKey, jwt.WithValidMethods([]string{
		jwt.SigningMethodRS256.Alg(),
	}))
}

func (a *JWTAuthorizer) publicKey(*jwt.Token) (any, error) {
	return a.PublicKey, nil
}
