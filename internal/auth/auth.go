package auth

import (
	"context"

	"bwawan.com/openuss/sdk/api"
)

type TokenSource interface {
	Token(ctx context.Context, audience string, scopes ...api.RequiredScope) (string, error)
}
