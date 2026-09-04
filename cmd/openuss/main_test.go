package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bwawan.com/openuss/internal/auth"
	"bwawan.com/openuss/internal/db"
	"bwawan.com/openuss/internal/dss"
	"bwawan.com/openuss/internal/logging/logtest"
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

func TestHandleShutdownReportsAFailedFlush(t *testing.T) {
	logger, rec := logtest.New()

	handleShutdown(func(context.Context) error {
		return errors.New("collector unreachable")
	}, logger)

	entry := rec.Find("tracing shutdown failed")
	require.NotNil(t, entry)
	require.Equal(t, "collector unreachable", entry["error"])
}

func TestHandleShutdownStaysQuietOnSuccess(t *testing.T) {
	logger, rec := logtest.New()
	handleShutdown(func(context.Context) error { return nil }, logger)
	require.Nil(t, rec.Find("tracing shutdown failed"))
}

func TestHandleShutdownGivesTheFlushALiveBudget(t *testing.T) {
	logger, _ := logtest.New()

	var (
		flushErr    error
		deadline    time.Time
		hasDeadline bool
	)

	handleShutdown(func(ctx context.Context) error {
		flushErr = ctx.Err()
		deadline, hasDeadline = ctx.Deadline()
		return nil
	}, logger)

	require.NoError(t, flushErr)
	require.True(t, hasDeadline)
	require.WithinDuration(t, time.Now().Add(flushTimeout), deadline, time.Second)
}

func TestNewPlanningHandler(t *testing.T) {
	os.Setenv("DSS_BASE_URL", "the-base-url")
	dummy, _ := auth.NewDummyOAuth("", "", nil)
	db := db.NewInMemoryDB()
	handler := newPlanningHandler(dummy, db)
	dss := handler.DSS.(*dss.DSS)
	require.Equal(t, http.DefaultClient, dss.Client)
	require.Equal(t, "the-base-url", dss.Host)
	require.Equal(t, "dss1.uss1.localutm", dss.Audience)
	require.Equal(t, dummy, dss.TokenSource)
	require.Equal(t, db, handler.DB)
}

func TestNewTokenSource(t *testing.T) {
	os.Setenv("OAUTH_ENDPOINT", "http://oauth.local")
	os.Setenv("OAUTH_SUB", "the-oauth-subject")
	source, err := newTokenSource()
	require.NoError(t, err)
	dummy := source.(*auth.DummyOAuth)
	require.Equal(t, "http://oauth.local", dummy.Endpoint.String())
	require.Equal(t, "the-oauth-subject", dummy.Subject)
}
