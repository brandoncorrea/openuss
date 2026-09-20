package dss_test

import (
	"net/http"
	"testing"
	"uuid"

	"bwawan.com/openuss/sdk/api"
	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/dss/dsstest"
	"bwawan.com/openuss/sdk/internal/wiretest"
	"bwawan.com/openuss/sdk/scdtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateIntentRespondsWithChangeResult(t *testing.T) {
	response := scdussv1.ChangeOperationalIntentReferenceResponse{}
	dssClient := dsstest.NewDSS(t, func(w http.ResponseWriter, r *http.Request) {
		api.WriteJSON(w, http.StatusOK, response)
	})
	reference := scdussv1.PutOperationalIntentReferenceParameters{}
	result, err := dssClient.PutOperationalIntentReference(t.Context(), scdtest.NewEntityID(), nil, reference)
	require.NoError(t, err)
	require.Equal(t, response, result)
}

func TestCreateIntentRequestParameters(t *testing.T) {
	entityID := scdtest.NewEntityID()
	dssClient := dsstest.NewDSS(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		token := "Bearer audience=dss.example.com&scopes=" + string(scdussv1.UtmStrategicCoordinationScope)
		assert.Equal(t, token, r.Header.Get("Authorization"))
		assert.Equal(t, "/dss/v1/operational_intent_references/"+string(entityID), r.RequestURI)
		api.WriteJSON(w, http.StatusOK, scdussv1.ChangeOperationalIntentReferenceResponse{})
	})
	reference := scdussv1.PutOperationalIntentReferenceParameters{}
	_, err := dssClient.PutOperationalIntentReference(t.Context(), entityID, nil, reference)
	require.NoError(t, err)
}

func TestUpdatesIntentWithSuppliedOVN(t *testing.T) {
	entityID := scdtest.NewEntityID()
	ovn := scdussv1.EntityOVN(uuid.New().String())
	dssClient := dsstest.NewDSS(t, func(w http.ResponseWriter, r *http.Request) {
		uri := "/dss/v1/operational_intent_references/" + string(entityID) + "/" + string(ovn)
		assert.Equal(t, uri, r.RequestURI)
		api.WriteJSON(w, http.StatusOK, scdussv1.ChangeOperationalIntentReferenceResponse{})
	})
	reference := scdussv1.PutOperationalIntentReferenceParameters{}
	_, err := dssClient.PutOperationalIntentReference(t.Context(), entityID, &ovn, reference)
	require.NoError(t, err)
}

func TestCreateIntentProducesErrorOnRequest(t *testing.T) {
	dssClient := dsstest.NewDSS(t, wiretest.AssertNotCalledHandler(t))
	reference := scdussv1.PutOperationalIntentReferenceParameters{}
	result, err := dssClient.PutOperationalIntentReference(nil, scdtest.NewEntityID(), nil, reference)
	require.Zero(t, result)
	require.Error(t, err)
}

func TestCreateIntentRespondsWithBadJSON(t *testing.T) {
	dssClient := dsstest.NewDSS(t, func(w http.ResponseWriter, r *http.Request) {
		api.WriteJSON(w, http.StatusOK, "{")
	})
	reference := scdussv1.PutOperationalIntentReferenceParameters{}
	result, err := dssClient.PutOperationalIntentReference(t.Context(), scdtest.NewEntityID(), nil, reference)
	require.Zero(t, result)
	require.Error(t, err)
}

func TestDeleteIntentRespondsWithChangeResult(t *testing.T) {
	ovn := scdussv1.EntityOVN(uuid.New().String())
	response := scdussv1.ChangeOperationalIntentReferenceResponse{}
	dssClient := dsstest.NewDSS(t, func(w http.ResponseWriter, r *http.Request) {
		api.WriteJSON(w, http.StatusOK, response)
	})
	result, err := dssClient.DeleteOperationalIntentReference(t.Context(), scdtest.NewEntityID(), ovn)
	require.NoError(t, err)
	require.Equal(t, response, result)
}

func TestDeleteIntentRequestParameters(t *testing.T) {
	entityID := string(scdtest.NewEntityID())
	ovn := uuid.New().String()
	dssClient := dsstest.NewDSS(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		token := "Bearer audience=dss.example.com&scopes=" + string(scdussv1.UtmStrategicCoordinationScope)
		assert.Equal(t, token, r.Header.Get("Authorization"))
		assert.Equal(t, "/dss/v1/operational_intent_references/"+entityID+"/"+ovn, r.RequestURI)
		api.WriteJSON(w, http.StatusOK, scdussv1.ChangeOperationalIntentReferenceResponse{})
	})
	_, err := dssClient.DeleteOperationalIntentReference(t.Context(), scdussv1.EntityID(entityID), scdussv1.EntityOVN(ovn))
	require.NoError(t, err)
}
