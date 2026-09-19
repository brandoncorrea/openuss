package httpclient_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"bwawan.com/openuss/internal/httpclient"
	"bwawan.com/openuss/internal/wiretest"
	"github.com/stretchr/testify/require"
)

func get(t *testing.T, client *http.Client, endpoint string) (httpclient.Response, error) {
	t.Helper()
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, endpoint, nil)
	require.NoError(t, err)
	return httpclient.New(client).Do(request)
}

func TestNewDefaultsToTimeoutClient(t *testing.T) {
	client := httpclient.New(nil)
	require.NotNil(t, client.HTTP)
	require.Equal(t, 10*time.Second, client.HTTP.Timeout)
}

func TestNewKeepsProvidedClient(t *testing.T) {
	httpClient := &http.Client{}
	client := httpclient.New(httpClient)
	require.Same(t, httpClient, client.HTTP)
}

func TestDoReturnsStatusAndBufferedBody(t *testing.T) {
	server := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "oh no!", http.StatusBadRequest)
	}))
	response, err := get(t, server.Client(), "http://example.com/foo")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.StatusCode)
	require.Equal(t, "oh no!\n", string(response.Body))
}

func TestDoReturnsEmptyBodyWhenNoneIsSent(t *testing.T) {
	server := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	response, err := get(t, server.Client(), "http://example.com/foo")
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, response.StatusCode)
	require.Empty(t, response.Body)
}

func TestDoReturnsTransportErrorUnwrapped(t *testing.T) {
	response, err := get(t, wiretest.NewErrorClient(errors.New("Boom!")), "http://example.com/foo")
	require.Zero(t, response)
	require.EqualError(t, err, `Get "http://example.com/foo": Boom!`)
}

type recordingBody struct {
	io.Reader
	drained bool
	closed  bool
}

func (r *recordingBody) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if err == io.EOF {
		r.drained = true
	}
	return n, err
}

func (r *recordingBody) Close() error {
	r.closed = true
	return nil
}

type cannedTransport struct {
	body io.ReadCloser
}

func (t *cannedTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK, Body: t.body}, nil
}

func cannedClient(body io.ReadCloser) *http.Client {
	return &http.Client{Transport: &cannedTransport{body: body}}
}

func TestDoDrainsAndClosesTheBody(t *testing.T) {
	body := &recordingBody{Reader: strings.NewReader("{}")}
	_, err := get(t, cannedClient(body), "http://example.com/foo")
	require.NoError(t, err)
	require.True(t, body.drained)
	require.True(t, body.closed)
}

func TestDoClosesTheBodyWhenReadingFails(t *testing.T) {
	body := &recordingBody{Reader: iotest.ErrReader(errors.New("Boom!"))}
	response, err := get(t, cannedClient(body), "http://example.com/foo")
	require.Zero(t, response)
	require.EqualError(t, err, "httpclient: reading response from http://example.com/foo: Boom!")
	require.True(t, body.closed)
}

func TestDoRefusesABodyOverTheCap(t *testing.T) {
	body := &recordingBody{Reader: strings.NewReader(strings.Repeat("x", httpclient.MaxBodyBytes+1))}
	response, err := get(t, cannedClient(body), "http://example.com/foo")
	require.Zero(t, response)
	require.EqualError(t, err, "httpclient: reading response from http://example.com/foo: body exceeds 1048576 bytes")
	require.True(t, body.closed)
}

func TestDoAcceptsABodyAtTheCap(t *testing.T) {
	body := &recordingBody{Reader: strings.NewReader(strings.Repeat("x", httpclient.MaxBodyBytes))}
	response, err := get(t, cannedClient(body), "http://example.com/foo")
	require.NoError(t, err)
	require.Len(t, response.Body, httpclient.MaxBodyBytes)
}
