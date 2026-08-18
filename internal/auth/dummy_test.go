package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"bwawan.com/openuss/internal/api"
	"github.com/stretchr/testify/require"
)

func newFakeServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(server.Close)
	return server
}

func fakeServerClient(t *testing.T) *http.Client {
	t.Helper()
	return newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		api.WriteJSON(w, http.StatusOK, map[string]any{
			"access_token": "the-token",
		})
	}).Client()
}

func requestToken(t *testing.T, server *httptest.Server, subject string, audience string, scopes ...string) (string, error) {
	t.Helper()
	auth, err := NewDummyOAuth(server.URL+"/token", subject, server.Client())
	require.NoError(t, err)
	return auth.Token(t.Context(), audience, scopes...)
}

func newToken(t *testing.T, audience string, scopes ...string) (string, error) {
	t.Helper()
	dummy, err := NewDummyOAuth("http://dummy/token", "foo_subject", fakeServerClient(t))
	require.NoError(t, err)
	return dummy.Token(t.Context(), audience, scopes...)
}

func TestNewDummyOAuthWithMalformedEndpoint(t *testing.T) {
	dummy, err := NewDummyOAuth("\t", "foo_subject", nil)
	require.Nil(t, dummy)
	require.ErrorContains(t, err, `auth: parsing token endpoint "\t": parse`)
}

func TestNewDummyOAuthEndpointMustBeAbsolute(t *testing.T) {
	dummy, err := NewDummyOAuth("/oauth/token", "foo_subject", nil)
	require.Nil(t, dummy)
	require.ErrorContains(t, err, `auth: token endpoint "/oauth/token" is not an absolute URL`)
}

func TestNewDummyOAuthEndpointMustHaveHost(t *testing.T) {
	dummy, err := NewDummyOAuth("http://", "foo_subject", nil)
	require.Nil(t, dummy)
	require.ErrorContains(t, err, `auth: token endpoint "http://" is not an absolute URL`)
}

func TestNewDummyOAuthMustHaveSubject(t *testing.T) {
	dummy, err := NewDummyOAuth("http://dummy/token", "\r\n\t ", nil)
	require.Nil(t, dummy)
	require.ErrorContains(t, err, "auth: subject is required")
}

func TestNewDummyOAuth(t *testing.T) {
	dummy, err := NewDummyOAuth("http://dummy/token", "foo_subject", nil)
	require.NoError(t, err)
	require.Equal(t, "foo_subject", dummy.subject)
	require.Equal(t, "http://dummy/token", dummy.endpoint.String())
	require.Equal(t, http.DefaultClient, dummy.client)
}

func TestNewDummyOAuthTrimsSubject(t *testing.T) {
	dummy, err := NewDummyOAuth("http://dummy/token", "  foo  subject  ", nil)
	require.NoError(t, err)
	require.Equal(t, "foo  subject", dummy.subject)
}

func TestNewDummyOAuthOverridesClient(t *testing.T) {
	client := fakeServerClient(t)
	dummy, err := NewDummyOAuth("http://dummy/token", "foo_subject", client)
	require.NoError(t, err)
	require.Equal(t, client, dummy.client)
}

func TestTokenWithNoAudience(t *testing.T) {
	token, err := newToken(t, "", "foo_scope")
	require.Equal(t, "", token)
	require.ErrorContains(t, err, "auth: audience is required")
}

func TestTokenWithNoScopes(t *testing.T) {
	token, err := newToken(t, "foo_audience")
	require.Equal(t, "", token)
	require.ErrorContains(t, err, "auth: at least one scope is required")
}

func TestMissingAudienceAndScopes(t *testing.T) {
	token, err := newToken(t, "")
	require.Equal(t, "", token)
	require.ErrorContains(t, err, "auth: audience is required")
}

func TestAudienceIsWhitespace(t *testing.T) {
	token, err := newToken(t, "\r\n\t ", "foo_scope")
	require.Equal(t, "", token)
	require.ErrorContains(t, err, "auth: audience is required")
}

func TestScopeIsBlank(t *testing.T) {
	token, err := newToken(t, "foo_audience", "\r\n\t ")
	require.Equal(t, "", token)
	require.ErrorContains(t, err, "auth: scope is empty")
}

func TestServerRequestErrors(t *testing.T) {
	server := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {})
	auth, err := NewDummyOAuth("http://error/token", "foo_subject", server.Client())
	require.NoError(t, err)
	token, err := auth.Token(t.Context(), "foo_audience", "foo_scope")
	require.Equal(t, "", token)
	require.ErrorContains(t, err, "auth: requesting token: Get \"http://error/token")
}

func TestServerReturnsNon2XX(t *testing.T) {
	server := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "oh no!", http.StatusBadRequest)
	})
	token, err := requestToken(t, server, "foo_subject", "foo_audience", "foo_scope")
	require.Equal(t, "", token)
	require.ErrorContains(t, err, "auth: token endpoint returned 400 Bad Request: oh no!\n")
}

func TestServerReturnsInvalidJson(t *testing.T) {
	server := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{"))
	})
	token, err := requestToken(t, server, "foo_subject", "foo_audience", "foo_scope")
	require.Equal(t, "", token)
	require.ErrorContains(t, err, "auth: decoding token response: unexpected EOF")
}

func TestServerReturnsUnreadableBody(t *testing.T) {
	server := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000")
		w.WriteHeader(http.StatusBadRequest)
	})
	token, err := requestToken(t, server, "foo_subject", "foo_audience", "foo_scope")
	require.Equal(t, "", token)
	require.ErrorContains(t, err, "auth: token endpoint returned 400 Bad Request: <unreadable body>")
}

func TestServerErrorBodyIsTruncated(t *testing.T) {
	server := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, strings.Repeat("x", maxErrorDetail*2), http.StatusBadRequest)
	})
	token, err := requestToken(t, server, "foo_subject", "foo_audience", "foo_scope")
	require.Equal(t, "", token)
	require.ErrorContains(t, err, strings.Repeat("x", maxErrorDetail))
	require.NotContains(t, err.Error(), strings.Repeat("x", maxErrorDetail+1))
}

func TestServerReturnsBlankToken(t *testing.T) {
	server := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		api.WriteJSON(w, http.StatusOK, map[string]any{
			"access_token": "\r\n\t ",
		})
	})
	token, err := requestToken(t, server, "foo_subject", "foo_audience", "foo_scope")
	require.Equal(t, "", token)
	require.ErrorContains(t, err, "auth: token response carried no access_token")
}

func TestServerReturnsToken(t *testing.T) {
	server := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		api.WriteJSON(w, http.StatusOK, map[string]any{
			"access_token": "the-token",
		})
	})
	token, err := requestToken(t, server, "foo_subject", "foo_audience", "foo_scope")
	require.NoError(t, err)
	require.Equal(t, "the-token", token)
}

func TestTokenRequestedWithQueryParams(t *testing.T) {
	var values url.Values
	server := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		values = r.URL.Query()
		api.WriteJSON(w, http.StatusOK, map[string]any{
			"access_token": "the-token",
		})
	})
	requestToken(t, server, "foo_subject", "foo_audience", "foo_scope", "bar_scope")
	require.Equal(t, "client_credentials", values.Get("grant_type"))
	require.Equal(t, "foo_scope bar_scope", values.Get("scope"))
	require.Equal(t, "foo_audience", values.Get("intended_audience"))
	require.Equal(t, "dummy", values.Get("issuer"))
	require.Equal(t, "foo_subject", values.Get("sub"))
}

func TestScopesGetTrimmed(t *testing.T) {
	var values url.Values
	server := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		values = r.URL.Query()
		api.WriteJSON(w, http.StatusOK, map[string]any{
			"access_token": "the-token",
		})
	})
	requestToken(t, server, "foo_subject", "foo_audience", " foo_scope ", " bar_scope ")
	require.Equal(t, "foo_scope bar_scope", values.Get("scope"))
}

func TestTokenWithMissingContext(t *testing.T) {
	auth, _ := NewDummyOAuth("http://dummy/token", "foo_subject", nil)
	var ctx context.Context
	token, err := auth.Token(ctx, "foo_audience", "foo_scope")
	require.Equal(t, "", token)
	require.ErrorContains(t, err, "auth: building token request: net/http: nil Context")
}
