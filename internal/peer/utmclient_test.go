package peer

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"bwawan.com/openuss/internal/api"
	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/auth"
	"bwawan.com/openuss/internal/utmclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOperationalIntentDetails(t *testing.T) {
	baseUrl := scdussv1.OperationalIntentUssBaseURL("http://uss1.localutm")
	entityId := scdussv1.EntityID(uuid.New().String())
	tokens := auth.NewInMemoryTokenSource()

	server := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Nil(t, r.TLS)
		assert.Equal(t, "uss1.localutm", r.Host)
		assert.Equal(t, "/uss/v1/operational_intents/"+string(entityId), r.RequestURI)
		assert.Equal(t, http.MethodGet, r.Method)

		token, err := tokens.Token(t.Context(), "uss1.localutm", scdussv1.UtmStrategicCoordinationScope)
		assert.NoError(t, err)
		assert.Equal(t, "Bearer "+token, r.Header.Get("Authorization"))

		api.WriteJSON(w, http.StatusOK, scdussv1.GetOperationalIntentDetailsResponse{
			OperationalIntent: scdussv1.OperationalIntent{
				Reference: scdussv1.OperationalIntentReference{
					Id: entityId,
				},
			},
		})
	}))

	client := New(utmclient.New(tokens, server.Client()))

	response, err := client.GetOperationalIntentDetails(t.Context(), baseUrl, entityId)
	require.NoError(t, err)
	require.Equal(t, entityId, response.OperationalIntent.Reference.Id)
}
