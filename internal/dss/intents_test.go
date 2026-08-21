package dss

import (
	"net/http"
	"testing"
	"uuid"

	"bwawan.com/openuss/internal/api"
	"bwawan.com/openuss/internal/api/scdussv1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateIntentRespondsWithChangeResult(t *testing.T) {
	entityId := scdussv1.EntityID(uuid.New().String())
	response := scdussv1.ChangeOperationalIntentReferenceResponse{}
	dss := newDss(t, func(w http.ResponseWriter, r *http.Request) {
		api.WriteJSON(w, http.StatusOK, response)
	})
	reference := scdussv1.PutOperationalIntentReferenceParameters{}
	result, err := dss.CreateOperationalIntentReference(t.Context(), entityId, reference)
	require.NoError(t, err)
	require.Equal(t, response, result)
}

func TestCreateIntentRequestParameters(t *testing.T) {
	entityId := scdussv1.EntityID(uuid.New().String())
	dss := newDss(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		token := "Bearer audience=dss.example.com&scopes=" + string(scdussv1.UtmStrategicCoordinationScope)
		assert.Equal(t, token, r.Header.Get("Authorization"))
		assert.Equal(t, "/dss/v1/operational_intent_references/"+string(entityId), r.RequestURI)
	})
	reference := scdussv1.PutOperationalIntentReferenceParameters{}
	dss.CreateOperationalIntentReference(t.Context(), entityId, reference)
}

func TestCreateIntentProducesErrorOnRequest(t *testing.T) {
	entityId := scdussv1.EntityID(uuid.New().String())
	dss := newDss(t, nil)
	reference := scdussv1.PutOperationalIntentReferenceParameters{}
	result, err := dss.CreateOperationalIntentReference(nil, entityId, reference)
	require.Zero(t, result)
	require.ErrorContains(t, err, "dss: failed to create request: net/http:")
}

func TestCreateIntentRespondsWithBadJson(t *testing.T) {
	entityId := scdussv1.EntityID(uuid.New().String())
	dss := newDss(t, func(w http.ResponseWriter, r *http.Request) {
		api.WriteJSON(w, http.StatusOK, "{")
	})
	reference := scdussv1.PutOperationalIntentReferenceParameters{}
	result, err := dss.CreateOperationalIntentReference(t.Context(), entityId, reference)
	require.Zero(t, result)
	require.ErrorContains(t, err, "dss: failed to parse response body: json:")
}
