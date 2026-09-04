package auth

import (
	"context"
	"strings"
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

func (source *InMemoryTokenSource) Token(_ context.Context, audience string, scopes ...string) (string, error) {
	token := "audience=" + audience + "&scopes=" + strings.Join(scopes, ",")
	return token, source.Error
}
