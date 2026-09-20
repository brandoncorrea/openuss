package auth

import (
	"context"
	"strings"

	"bwawan.com/openuss/sdk/api"
	"bwawan.com/openuss/sdk/internal/util"
)

type InMemoryTokenSource struct {
	err error
}

func NewInMemoryTokenSource() *InMemoryTokenSource {
	return &InMemoryTokenSource{}
}

func NewInMemoryErrorTokenSource(err error) *InMemoryTokenSource {
	return &InMemoryTokenSource{err: err}
}

func (t *InMemoryTokenSource) Token(_ context.Context, audience string, scopes ...api.RequiredScope) (string, error) {
	scope := strings.Join(util.Map(scopes, util.TrimString), ",")
	token := "audience=" + audience + "&scopes=" + scope
	return token, t.err
}
