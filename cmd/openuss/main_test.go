package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListenAddr(t *testing.T) {
	tests := []struct {
		name string
		port string
		want string
	}{
		{"defaults to port 80", "", ":80"},
		{"port 8080", "8080", ":8080"},
		{"port 3000", "3000", ":3000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PORT", tt.port)
			assert.Equal(t, tt.want, ResolveAddress())
		})
	}
}
