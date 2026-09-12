package auth

import (
	"context"
	"strings"

	"bwawan.com/openuss/internal/api"
	"bwawan.com/openuss/internal/util"
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

func (source *InMemoryTokenSource) Token(_ context.Context, audience string, scopes ...api.RequiredScope) (string, error) {
	scope := strings.Join(util.Map(scopes, util.TrimString), ",")
	token := "audience=" + audience + "&scopes=" + scope
	return token, source.Error
}
