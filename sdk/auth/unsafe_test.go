package auth_test

import (
	"testing"

	"bwawan.com/openuss/sdk/auth"
	"github.com/stretchr/testify/require"
)

func TestAcceptsRequestsWithoutBearerToken(t *testing.T) {
	authorizer := auth.NewUnsafeAuthorizer()
	request := emptyRequest()
	result := authorizeGetDetails(authorizer, request)
	require.NoError(t, result.Error)
}
