package server

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"bwawan.com/openuss/internal/logging/logtest"
)

func listenForTest(t *testing.T) (*Server, *logtest.Recorder) {
	t.Helper()

	logger, rec := logtest.New()

	server, err := Listen("127.0.0.1:0", http.NewServeMux(), logger)
	require.NoError(t, err)
	t.Cleanup(func() { server.Close() })

	return server, rec
}

func TestListenReportsResolvedAddress(t *testing.T) {
	server, _ := listenForTest(t)
	require.NotEqual(t, "127.0.0.1:0", server.Addr().String())
	require.NotZero(t, server.Addr().(*net.TCPAddr).Port)
}

func TestListenFailsOnAddressInUse(t *testing.T) {
	taken, _ := listenForTest(t)
	logger, _ := logtest.New()

	_, err := Listen(taken.Addr().String(), http.NewServeMux(), logger)
	require.Error(t, err)
}

func TestListenUsesProvidedHandler(t *testing.T) {
	handler := http.NewServeMux()
	logger, _ := logtest.New()

	server, err := Listen("127.0.0.1:0", handler, logger)
	require.NoError(t, err)
	t.Cleanup(func() { server.Close() })

	require.Same(t, handler, server.http.Handler)
}

func TestListenBoundsEveryTimeout(t *testing.T) {
	server, _ := listenForTest(t)
	require.Equal(t, 15*time.Second, server.http.ReadHeaderTimeout)
	require.Equal(t, 15*time.Second, server.http.ReadTimeout)
	require.Equal(t, 10*time.Second, server.http.WriteTimeout)
	require.Equal(t, 30*time.Second, server.http.IdleTimeout)
	require.Equal(t, defaultShutdownTimeout, server.shutdownTimeout)
}

func TestRunServesThenShutsDownOnCancel(t *testing.T) {
	server, rec := listenForTest(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- server.Run(ctx) }()

	resp, err := http.Get("http://" + server.Addr().String() + "/")
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	cancel()

	err = awaitRun(t, done)
	require.NoError(t, err)
	entry := rec.Find("draining connections")
	require.NotNil(t, entry, "expected a drain log line")
	require.EqualValues(t, defaultShutdownTimeout, entry["timeout"])
}

// A handler still waiting on the DSS when SIGTERM lands outlives the drain
// deadline. Run must report that rather than let main exit 0.
func TestRunReportsFailedDrain(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	t.Cleanup(func() { close(release) })

	mux := http.NewServeMux()
	mux.HandleFunc("GET /slow", func(http.ResponseWriter, *http.Request) {
		close(entered)
		<-release
	})

	logger, rec := logtest.New()

	server, err := Listen("127.0.0.1:0", mux, logger)
	require.NoError(t, err)
	t.Cleanup(func() { server.Close() })
	server.shutdownTimeout = 50 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- server.Run(ctx) }()

	go http.Get("http://" + server.Addr().String() + "/slow")
	<-entered

	cancel()

	err = awaitRun(t, done)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.NotNil(t, rec.Find("drain did not complete"), "expected a failed-drain log line")
}

func awaitRun(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		require.Fail(t, "Run did not return")
		return nil
	}
}
