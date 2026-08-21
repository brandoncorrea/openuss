package testutil

import (
	"context"
	"strings"
)

type FakeTokenSource struct {
	Error error
}

func (source *FakeTokenSource) Token(_ context.Context, audience string, scopes ...string) (string, error) {
	return EncodeFakeToken(audience, scopes), source.Error
}

func EncodeFakeToken(audience string, scopes []string) string {
	return "audience=" + audience + "&scopes=" + strings.Join(scopes, ",")
}
