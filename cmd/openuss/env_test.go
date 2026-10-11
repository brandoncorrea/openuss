package main_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	main "bwawan.com/openuss/cmd/openuss"
)

func getenvFrom(variables map[string]string) func(string) string {
	return func(key string) string {
		return variables[key]
	}
}

func TestParseEnvironmentReadsEverySetting(t *testing.T) {
	env := main.ParseEnvironment(getenvFrom(map[string]string{
		"AUTHORIZER_IMPL":       "none",
		"DSS_BASE_URL":          "http://dss.localutm",
		"DSS_IMPL":              "memory",
		"OAUTH_ENDPOINT":        "http://oauth.localutm/token",
		"OAUTH_PUBLIC_KEY_FILE": "path-to-auth2.pem",
		"OAUTH_SUB":             "uss5",
		"PORT":                  "8123",
		"TOKEN_IMPL":            "memory",
		"USS_BASE_URL":          "http://openuss.localutm",
	}))

	require.Equal(t, main.Environment{
		DSSBaseURL:         "http://dss.localutm",
		InMemoryDSS:        true,
		InMemoryTokens:     true,
		ListenAddress:      ":8123",
		OAuthEndpoint:      "http://oauth.localutm/token",
		OAuthPublicKeyFile: "path-to-auth2.pem",
		OAuthSubject:       "uss5",
		UnsafeAuthorizer:   true,
		USSBaseURL:         "http://openuss.localutm",
	}, env)
}

func TestParseEnvironmentUsesRealImplementationsForUnrecognizedValues(t *testing.T) {
	env := main.ParseEnvironment(getenvFrom(map[string]string{
		"AUTHORIZER_IMPL": "blah",
		"DSS_IMPL":        "blah",
		"TOKEN_IMPL":      "blah",
	}))

	require.False(t, env.InMemoryDSS)
	require.False(t, env.InMemoryTokens)
	require.False(t, env.UnsafeAuthorizer)
}

func TestParseEnvironmentListensOn8080AndUsesRealImplementationsByDefault(t *testing.T) {
	env := main.ParseEnvironment(getenvFrom(nil))
	require.Equal(t, main.Environment{ListenAddress: ":8080"}, env)
}
