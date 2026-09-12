package utmclient

import (
	"encoding/json/v2"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bwawan.com/openuss/internal/api"
	"bwawan.com/openuss/internal/auth"
	"bwawan.com/openuss/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newClient(t *testing.T, handler http.HandlerFunc) *Client {
	server := httptest.NewTestServer(t, handler)
	return New(auth.NewInMemoryTokenSource(), server.Client())
}

func TestNewDefaultsToTimeoutClient(t *testing.T) {
	client := New(auth.NewInMemoryTokenSource(), nil)
	require.NotNil(t, client.HTTP)
	require.Equal(t, 10*time.Second, client.HTTP.Timeout)
}

func TestNewKeepsProvidedClient(t *testing.T) {
	httpClient := &http.Client{}
	client := New(auth.NewInMemoryTokenSource(), httpClient)
	require.Same(t, httpClient, client.HTTP)
}

func requireGetSuccess(
	t *testing.T,
	endpoint string,
	handler func(http.ResponseWriter, *http.Request),
	scopes ...api.RequiredScope,
) {
	t.Helper()
	_, err := newClient(t, handler).Get(t.Context(), endpoint, scopes...)
	require.NoError(t, err)
}

func requirePostSuccess(
	t *testing.T,
	endpoint string,
	body any,
	handler func(http.ResponseWriter, *http.Request),
	scopes ...api.RequiredScope,
) {
	t.Helper()
	_, err := newClient(t, handler).Post(t.Context(), endpoint, body, scopes...)
	require.NoError(t, err)
}

func TestDoRequestOptions(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Nil(t, r.TLS)
		assert.Equal(t, "dss.example.com", r.Host)
		assert.Equal(t, "/foo", r.RequestURI)
		assert.Equal(t, http.MethodGet, r.Method)
	}
	requireGetSuccess(t, "http://dss.example.com/foo", handler)
}

func TestDoUsesHostnameAsAudience(t *testing.T) {
	tokenSource := auth.NewInMemoryTokenSource()
	handler := func(w http.ResponseWriter, r *http.Request) {
		token, _ := tokenSource.Token(nil, "dss.example.com", "scope-1", "scope-2")
		assert.Equal(t, "Bearer "+token, r.Header.Get("Authorization"))
	}
	requireGetSuccess(t, "http://dss.example.com", handler, "scope-1", "scope-2")
}

func TestDoStripsPortFromAudience(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "peer.example.com:8080", r.Host)
		assert.Equal(t, "Bearer audience=peer.example.com&scopes=scope-1", r.Header.Get("Authorization"))
	}
	requireGetSuccess(t, "http://peer.example.com:8080/foo", handler, "scope-1")
}

func TestDoDropsAudienceToLowerCase(t *testing.T) {
	endpoint := "https://Openuss.uss5.localutm:8080/uss/v1/operational_intents/x"
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Openuss.uss5.localutm:8080", r.Host)
		assert.Equal(t, "Bearer audience=openuss.uss5.localutm&scopes=scope-1", r.Header.Get("Authorization"))
	}
	requireGetSuccess(t, endpoint, handler, "scope-1")
}

func TestDoSendsBody(t *testing.T) {
	body := map[string]any{
		"foo": "bar",
		"baz": "buzz",
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		var requestedBody map[string]any
		assert.NoError(t, json.UnmarshalRead(r.Body, &requestedBody))
		assert.Equal(t, body, requestedBody)
	}
	requirePostSuccess(t, "http://dss.example.com", body, handler)
}

func TestDoSendsNoBodyWhenNil(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Empty(t, body)
	}
	requireGetSuccess(t, "http://dss.example.com", handler)
}

func TestDoFailsOnUnparsableUrl(t *testing.T) {
	client := newClient(t, testutil.AssertNotCalledHandler(t))
	response, err := client.Get(t.Context(), "http://%zz")
	require.Nil(t, response)
	require.ErrorContains(t, err, "utmclient: failed to parse url")
}

func TestDoFailsOnUrlWithoutHostname(t *testing.T) {
	client := newClient(t, testutil.AssertNotCalledHandler(t))
	response, err := client.Get(t.Context(), "/foo")
	require.Nil(t, response)
	require.ErrorContains(t, err, `utmclient: url "/foo" has no hostname`)
}

func TestDoFailsToCreateNewRequest(t *testing.T) {
	client := newClient(t, testutil.AssertNotCalledHandler(t))
	response, err := client.Get(nil, "http://dss.example.com")
	require.Nil(t, response)
	require.ErrorContains(t, err, "utmclient: failed to create request: net/http:")
}

func TestDoFailsToProduceToken(t *testing.T) {
	client := newClient(t, testutil.AssertNotCalledHandler(t))
	client.TokenSource = auth.NewInMemoryErrorTokenSource(errors.New("Boom!"))
	response, err := client.Get(t.Context(), "http://dss.example.com")
	require.Nil(t, response)
	require.ErrorContains(t, err, "utmclient: failed to acquire auth token: Boom!")
}

func TestDoFailsToMarshalBody(t *testing.T) {
	client := newClient(t, testutil.AssertNotCalledHandler(t))
	unmarshallable := make(chan int)
	response, err := client.Post(t.Context(), "http://dss.example.com", unmarshallable)
	require.Nil(t, response)
	require.ErrorContains(t, err, "utmclient: failed to encode request body: json:")
}

func TestDoReturnsTransportError(t *testing.T) {
	client := newClient(t, testutil.AssertNotCalledHandler(t))
	client.HTTP = testutil.NewErrorClient(errors.New("Boom!"))
	response, err := client.Get(t.Context(), "http://dss.example.com")
	require.Nil(t, response)
	require.ErrorContains(t, err, "utmclient: failed to make request: Get")
}
