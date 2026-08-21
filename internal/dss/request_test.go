package dss

import (
	"encoding/json/v2"
	"errors"
	"net/http"
	"testing"

	"bwawan.com/openuss/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeRequestOptions(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Nil(t, r.TLS)
		assert.Equal(t, "dss.example.com", r.Host)
		assert.Equal(t, "/foo", r.RequestURI)
		assert.Equal(t, http.MethodGet, r.Method)
	}
	newDss(t, handler).MakeRequest(t.Context(), http.MethodGet, "/foo", map[string]any{})
}

func TestMakeRequestGeneratesPropertToken(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		expected := "Bearer " + testutil.EncodeFakeToken("dss.example.com", []string{"scope-1", "scope-2"})
		require.Equal(t, expected, r.Header.Get("Authorization"))
	}
	newDss(t, handler).MakeRequest(
		t.Context(),
		http.MethodGet,
		"",
		map[string]any{},
		"scope-1",
		"scope-2")
}

func TestMakeRequestBody(t *testing.T) {
	body := map[string]any{
		"foo": "bar",
		"baz": "buzz",
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		var requestedBody map[string]any
		assert.NoError(t, json.UnmarshalRead(r.Body, &requestedBody))
		assert.Equal(t, body, requestedBody)
	}
	newDss(t, handler).MakeRequest(t.Context(), http.MethodGet, "", body)
}

func TestMakeRequestFailsToCreateNewRequest(t *testing.T) {
	dss := newDss(t, assertNotCalledHandler(t))
	response, err := dss.MakeRequest(nil, http.MethodGet, "", map[string]any{})
	require.Nil(t, response)
	require.ErrorContains(t, err, "dss: failed to create request: net/http:")
}

func TestMakeRequestFailsToProduceToken(t *testing.T) {
	dss := newDss(t, assertNotCalledHandler(t))
	dss.TokenSource = &testutil.FakeTokenSource{
		Error: errors.New("Boom!"),
	}
	response, err := dss.MakeRequest(t.Context(), http.MethodGet, "", map[string]any{})
	require.Nil(t, response)
	require.ErrorContains(t, err, "dss: failed to acquire auth token: Boom!")
}

func TestMakeRequestFailsToMarshalBody(t *testing.T) {
	dss := newDss(t, assertNotCalledHandler(t))
	badBody := make(chan int)
	response, err := dss.MakeRequest(t.Context(), http.MethodGet, "", badBody)
	require.Nil(t, response)
	require.ErrorContains(t, err, "dss: failed to encode request body: json:")
}

func TestMakeRequestReturnsTransportError(t *testing.T) {
	boom := errors.New("Boom!")
	dss := newDss(t, assertNotCalledHandler(t))
	dss.Client = testutil.NewErrorClient(boom)
	response, err := dss.MakeRequest(t.Context(), http.MethodGet, "", map[string]any{})
	require.Nil(t, response)
	require.ErrorContains(t, err, "dss: failed to make request: Get")
}
