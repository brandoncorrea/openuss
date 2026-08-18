package auth

import "context"

type TokenSource interface {
	Token(ctx context.Context, audience string, scopes ...string) (string, error)
}
