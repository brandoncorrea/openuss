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
		"DSS_BASE_URL":   "http://dss.localutm",
		"DSS_IMPL":       "memory",
		"OAUTH_ENDPOINT": "http://oauth.localutm/token",
		"OAUTH_SUB":      "uss5",
		"PORT":           "8123",
		"TOKEN_IMPL":     "memory",
		"USS_BASE_URL":   "http://openuss.localutm",
	}))

	require.Equal(t, main.Environment{
		DSSBaseURL:     "http://dss.localutm",
		InMemoryDSS:    true,
		InMemoryTokens: true,
		ListenAddress:  ":8123",
		OAuthEndpoint:  "http://oauth.localutm/token",
		OAuthSubject:   "uss5",
		USSBaseURL:     "http://openuss.localutm",
	}, env)
}

func TestParseEnvironmentUsesRealImplementationsForAnyValueButMemory(t *testing.T) {
	env := main.ParseEnvironment(getenvFrom(map[string]string{
		"DSS_IMPL":   "blah",
		"TOKEN_IMPL": "blah",
	}))

	require.False(t, env.InMemoryDSS)
	require.False(t, env.InMemoryTokens)
}

func TestParseEnvironmentListensOn8080AndUsesRealImplementationsByDefault(t *testing.T) {
	env := main.ParseEnvironment(getenvFrom(nil))

	require.Equal(t, main.Environment{ListenAddress: ":8080"}, env)
}
