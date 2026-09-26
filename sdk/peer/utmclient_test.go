package peer_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bwawan.com/openuss/sdk/api"
	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/auth"
	"bwawan.com/openuss/sdk/peer"
	"bwawan.com/openuss/sdk/scdtest"
	"bwawan.com/openuss/sdk/utmclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOperationalIntentDetails(t *testing.T) {
	baseURL := scdussv1.OperationalIntentUssBaseURL("http://uss1.localutm")
	entityID := scdtest.NewEntityID()
	tokens := auth.NewInMemoryTokenSource()

	server := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Nil(t, r.TLS)
		assert.Equal(t, "uss1.localutm", r.Host)
		assert.Equal(t, "/uss/v1/operational_intents/"+string(entityID), r.RequestURI)
		assert.Equal(t, http.MethodGet, r.Method)

		token, err := tokens.Token(t.Context(), "uss1.localutm", scdussv1.UtmStrategicCoordinationScope)
		assert.NoError(t, err)
		assert.Equal(t, "Bearer "+token, r.Header.Get("Authorization"))

		api.WriteJSON(w, http.StatusOK, scdussv1.GetOperationalIntentDetailsResponse{
			OperationalIntent: scdussv1.OperationalIntent{
				Reference: scdussv1.OperationalIntentReference{
					Id: entityID,
				},
			},
		})
	}))

	client := peer.New(utmclient.New(tokens, server.Client()))

	response, err := client.GetOperationalIntentDetails(t.Context(), baseURL, entityID)
	require.NoError(t, err)
	require.Equal(t, entityID, response.OperationalIntent.Reference.Id)
}

func TestNotifyOperationalIntentDetails(t *testing.T) {
	baseURL := scdussv1.SubscriptionUssBaseURL("http://uss1.localutm")
	entityID := scdtest.NewEntityID()
	details := scdussv1.PutOperationalIntentDetailsParameters{
		OperationalIntentId: entityID,
	}
	tokens := auth.NewInMemoryTokenSource()

	server := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Nil(t, r.TLS)
		assert.Equal(t, "uss1.localutm", r.Host)
		assert.Equal(t, "/uss/v1/operational_intents", r.RequestURI)
		assert.Equal(t, http.MethodPost, r.Method)

		token, err := tokens.Token(t.Context(), "uss1.localutm", scdussv1.UtmStrategicCoordinationScope)
		assert.NoError(t, err)
		assert.Equal(t, "Bearer "+token, r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusNoContent)
	}))

	client := peer.New(utmclient.New(tokens, server.Client()))

	err := client.NotifyOperationalIntentDetails(t.Context(), baseURL, details)
	require.NoError(t, err)
}
