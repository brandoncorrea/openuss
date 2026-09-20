package auth

import (
	"context"
	"strings"

	"bwawan.com/openuss/internal/util"
	"bwawan.com/openuss/sdk/api"
)

type InMemoryTokenSource struct {
	Error error
}

func NewInMemoryTokenSource() *InMemoryTokenSource {
	return &InMemoryTokenSource{}
}

func NewInMemoryErrorTokenSource(err error) *InMemoryTokenSource {
	return &InMemoryTokenSource{Error: err}
}

func (t *InMemoryTokenSource) Token(_ context.Context, audience string, scopes ...api.RequiredScope) (string, error) {
	scope := strings.Join(util.Map(scopes, util.TrimString), ",")
	token := "audience=" + audience + "&scopes=" + scope
	return token, t.Error
}
