package utmclient_test

import (
	"encoding/json/v2"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/sdk/api"
	"bwawan.com/openuss/sdk/auth"
	"bwawan.com/openuss/sdk/httpclient"
	"bwawan.com/openuss/sdk/internal/wiretest"
	"bwawan.com/openuss/sdk/utmclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newClient(t *testing.T, handler http.HandlerFunc) *utmclient.Client {
	server := httptest.NewTestServer(t, handler)
	return utmclient.New(auth.NewInMemoryTokenSource(), server.Client())
}

func TestNewDefaultsHTTPClient(t *testing.T) {
	client := utmclient.New(auth.NewInMemoryTokenSource(), nil)
	require.Equal(t, httpclient.DefaultTimeout, client.HTTP.HTTP.Timeout)
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
		token, _ := tokenSource.Token(t.Context(), "dss.example.com", "scope-1", "scope-2")
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

func TestDoFailsOnUnparsableURL(t *testing.T) {
	client := newClient(t, wiretest.AssertNotCalledHandler(t))
	response, err := client.Get(t.Context(), "http://%zz")
	require.Zero(t, response)
	require.ErrorContains(t, err, "utmclient: failed to parse url")
}

func TestDoFailsOnURLWithoutHostname(t *testing.T) {
	client := newClient(t, wiretest.AssertNotCalledHandler(t))
	response, err := client.Get(t.Context(), "/foo")
	require.Zero(t, response)
	require.ErrorContains(t, err, `utmclient: url "/foo" has no hostname`)
}

func TestDoFailsToCreateNewRequest(t *testing.T) {
	client := newClient(t, wiretest.AssertNotCalledHandler(t))
	response, err := client.Get(nil, "http://dss.example.com")
	require.Zero(t, response)
	require.ErrorContains(t, err, "utmclient: failed to create request: net/http:")
}

func TestDoFailsToProduceToken(t *testing.T) {
	client := newClient(t, wiretest.AssertNotCalledHandler(t))
	client.TokenSource = auth.NewInMemoryErrorTokenSource(errors.New("Boom!"))
	response, err := client.Get(t.Context(), "http://dss.example.com")
	require.Zero(t, response)
	require.ErrorContains(t, err, "utmclient: failed to acquire auth token: Boom!")
}

func TestDoFailsToMarshalBody(t *testing.T) {
	client := newClient(t, wiretest.AssertNotCalledHandler(t))
	unmarshallable := make(chan int)
	response, err := client.Post(t.Context(), "http://dss.example.com", unmarshallable)
	require.Zero(t, response)
	require.ErrorContains(t, err, "utmclient: failed to encode request body: json:")
}

func TestDoReturnsTransportError(t *testing.T) {
	client := newClient(t, wiretest.AssertNotCalledHandler(t))
	client.HTTP = httpclient.New(wiretest.NewErrorClient(errors.New("Boom!")))
	response, err := client.Get(t.Context(), "http://dss.example.com")
	require.Zero(t, response)
	require.ErrorContains(t, err, `Get "http://dss.example.com": Boom!`)
}

func TestDoReturnsStatusAndBufferedBody(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		api.WriteJSON(w, http.StatusConflict, map[string]string{
			"foo": "bar",
		})
	}
	response, err := newClient(t, handler).Get(t.Context(), "http://dss.example.com")
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, response.StatusCode)
	require.Equal(t, "{\"foo\":\"bar\"}\n", string(response.Body))
}

func TestDoReturnsEmptyBodyWhenNoneIsSent(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
	response, err := newClient(t, handler).Get(t.Context(), "http://dss.example.com")
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, response.StatusCode)
	require.Empty(t, response.Body)
}
