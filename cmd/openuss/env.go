package main

import (
	"cmp"
	"os"

	"github.com/joho/godotenv"
)

type Environment struct {
	ListenAddress  string
	USSBaseURL     string
	InMemoryDSS    bool
	DSSBaseURL     string
	InMemoryTokens bool
	OAuthEndpoint  string
	OAuthSubject   string
}

func ParseEnvironment(getenv func(string) string) Environment {
	return Environment{
		DSSBaseURL:     getenv("DSS_BASE_URL"),
		InMemoryDSS:    getenv("DSS_IMPL") == "memory",
		InMemoryTokens: getenv("TOKEN_IMPL") == "memory",
		ListenAddress:  ":" + cmp.Or(getenv("PORT"), "8080"),
		OAuthEndpoint:  getenv("OAUTH_ENDPOINT"),
		OAuthSubject:   getenv("OAUTH_SUB"),
		USSBaseURL:     getenv("USS_BASE_URL"),
	}
}

func loadDotEnv() error {
	err := godotenv.Load()
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
