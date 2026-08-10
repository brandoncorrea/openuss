package server

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func listenForTest(t *testing.T) *Server {
	t.Helper()

	server, err := Listen("127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { server.Close() })

	return server
}

func TestListenReportsResolvedAddress(t *testing.T) {
	server := listenForTest(t)
	require.NotEqual(t, "127.0.0.1:0", server.Addr().String())
	require.NotZero(t, server.Addr().(*net.TCPAddr).Port)
}

func TestListenFailsOnAddressInUse(t *testing.T) {
	taken := listenForTest(t)
	_, err := Listen(taken.Addr().String())
	require.Error(t, err)
}

func TestListenBoundsEveryTimeout(t *testing.T) {
	server := listenForTest(t)
	require.IsType(t, &http.ServeMux{}, server.http.Handler)
	require.Equal(t, 15*time.Second, server.http.ReadHeaderTimeout)
	require.Equal(t, 15*time.Second, server.http.ReadTimeout)
	require.Equal(t, 10*time.Second, server.http.WriteTimeout)
	require.Equal(t, 30*time.Second, server.http.IdleTimeout)
}

func TestRunServesThenShutsDownOnCancel(t *testing.T) {
	server := listenForTest(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- server.Run(ctx) }()

	resp, err := http.Get("http://" + server.Addr().String() + "/")
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after context cancellation")
	}
}
