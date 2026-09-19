package server_test

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"bwawan.com/openuss/internal/logging/logtest"
	"bwawan.com/openuss/internal/server"
)

func listenForTest(t *testing.T) (*server.Server, *logtest.Recorder) {
	t.Helper()

	logger, rec := logtest.New()

	srv, err := server.Listen("127.0.0.1:0", http.NewServeMux(), logger)
	require.NoError(t, err)
	t.Cleanup(func() { srv.Close() })

	return srv, rec
}

func TestListenReportsResolvedAddress(t *testing.T) {
	srv, _ := listenForTest(t)
	require.NotEqual(t, "127.0.0.1:0", srv.Addr().String())
	require.NotZero(t, srv.Addr().(*net.TCPAddr).Port)
}

func TestListenFailsOnAddressInUse(t *testing.T) {
	taken, _ := listenForTest(t)
	logger, _ := logtest.New()

	_, err := server.Listen(taken.Addr().String(), http.NewServeMux(), logger)
	require.Error(t, err)
}

func TestListenUsesProvidedHandler(t *testing.T) {
	handler := http.NewServeMux()
	logger, _ := logtest.New()

	srv, err := server.Listen("127.0.0.1:0", handler, logger)
	require.NoError(t, err)
	t.Cleanup(func() { srv.Close() })

	require.Same(t, handler, srv.HTTP.Handler)
}

func TestListenBoundsEveryTimeout(t *testing.T) {
	srv, _ := listenForTest(t)
	require.Equal(t, 15*time.Second, srv.HTTP.ReadHeaderTimeout)
	require.Equal(t, 15*time.Second, srv.HTTP.ReadTimeout)
	require.Equal(t, 10*time.Second, srv.HTTP.WriteTimeout)
	require.Equal(t, 30*time.Second, srv.HTTP.IdleTimeout)
	require.Equal(t, server.DefaultShutdownTimeout, srv.ShutdownTimeout)
}

func TestRunServesThenShutsDownOnCancel(t *testing.T) {
	srv, rec := listenForTest(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- srv.Run(ctx) }()

	resp, err := http.Get("http://" + srv.Addr().String() + "/")
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	cancel()

	err = awaitRun(t, done)
	require.NoError(t, err)
	entry := rec.Find("draining connections")
	require.NotNil(t, entry, "expected a drain log line")
	require.EqualValues(t, server.DefaultShutdownTimeout, entry["timeout"])
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

	srv, err := server.Listen("127.0.0.1:0", mux, logger)
	require.NoError(t, err)
	t.Cleanup(func() { srv.Close() })
	srv.ShutdownTimeout = 50 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- srv.Run(ctx) }()

	go http.Get("http://" + srv.Addr().String() + "/slow")
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
