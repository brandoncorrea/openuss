package main

import (
	"cmp"
	"os"

	"github.com/joho/godotenv"
)

type Environment struct {
	DSSBaseURL         string
	InMemoryDSS        bool
	InMemoryTokens     bool
	ListenAddress      string
	OAuthEndpoint      string
	OAuthPublicKeyFile string
	OAuthSubject       string
	UnsafeAuthorizer   bool
	USSBaseURL         string
}

func ParseEnvironment(getenv func(string) string) Environment {
	return Environment{
		DSSBaseURL:         getenv("DSS_BASE_URL"),
		InMemoryDSS:        getenv("DSS_IMPL") == "memory",
		InMemoryTokens:     getenv("TOKEN_IMPL") == "memory",
		ListenAddress:      ":" + cmp.Or(getenv("PORT"), "8080"),
		OAuthEndpoint:      getenv("OAUTH_ENDPOINT"),
		OAuthPublicKeyFile: getenv("OAUTH_PUBLIC_KEY_FILE"),
		OAuthSubject:       getenv("OAUTH_SUB"),
		UnsafeAuthorizer:   getenv("AUTHORIZER_IMPL") == "none",
		USSBaseURL:         getenv("USS_BASE_URL"),
	}
}

func loadDotEnv() error {
	err := godotenv.Load()
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
