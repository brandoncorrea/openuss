package auth

import (
	"context"

	"bwawan.com/openuss/internal/api"
)

type TokenSource interface {
	Token(ctx context.Context, audience string, scopes ...api.RequiredScope) (string, error)
}
