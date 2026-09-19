package dsstest

import (
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"uuid"

	"bwawan.com/openuss/internal/api"
	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/auth"
	"bwawan.com/openuss/internal/dss"
	"bwawan.com/openuss/internal/util"
	"bwawan.com/openuss/internal/utmclient"
)

func NewDSS(t *testing.T, handler http.HandlerFunc) *dss.DSS {
	t.Helper()
	server := httptest.NewTestServer(t, http.HandlerFunc(handler))
	client := utmclient.New(auth.NewInMemoryTokenSource(), server.Client())
	return dss.New("http://dss.example.com", client)
}

func NewPeerHandler(peers []scdussv1.OperationalIntent) http.HandlerFunc {
	handleDSS := dssHandlerFromPeers(peers)
	handleUSS := ussHandlerFromPeers(peers)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Host == "dss.localutm" {
			handleDSS(w, r)
		} else {
			handleUSS(w, r)
		}
	}
}

func ussHandlerFromPeers(peers []scdussv1.OperationalIntent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		index := slices.IndexFunc(peers, func(intent scdussv1.OperationalIntent) bool {
			host := strings.TrimPrefix(string(intent.Reference.UssBaseUrl), "http://")
			return host == r.Host && string(intent.Reference.Id) == uriEntityID(r.RequestURI)
		})

		if index >= 0 {
			api.WriteJSON(w, http.StatusOK, scdussv1.GetOperationalIntentDetailsResponse{
				OperationalIntent: peers[index],
			})
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func dssHandlerFromPeers(peers []scdussv1.OperationalIntent) http.HandlerFunc {
	references := util.Map(peers, func(intent scdussv1.OperationalIntent) scdussv1.OperationalIntentReference {
		return intent.Reference
	})

	return func(w http.ResponseWriter, r *http.Request) {
		var putParams scdussv1.PutOperationalIntentReferenceParameters
		json.UnmarshalRead(r.Body, &putParams)

		key := []scdussv1.EntityOVN{}
		if putParams.Key != nil {
			key = *putParams.Key
		}
		missing := util.Remove(references, func(reference scdussv1.OperationalIntentReference) bool {
			return slices.Contains(key, *reference.Ovn)
		})

		if len(missing) > 0 {
			masked := util.Map(missing, func(reference scdussv1.OperationalIntentReference) scdussv1.OperationalIntentReference {
				reference.Ovn = new(scdussv1.EntityOVN("blah"))
				return reference
			})
			api.WriteJSON(w, http.StatusConflict, scdussv1.AirspaceConflictResponse{
				MissingOperationalIntents: &masked,
			})
		} else {
			api.WriteJSON(w, http.StatusOK, scdussv1.ChangeOperationalIntentReferenceResponse{
				OperationalIntentReference: scdussv1.OperationalIntentReference{
					Id:         scdussv1.EntityID(uriEntityID(r.RequestURI)),
					Ovn:        new(scdussv1.EntityOVN(uuid.New().String())),
					UssBaseUrl: putParams.UssBaseUrl,
				},
			})
		}
	}
}

func uriEntityID(uri string) string {
	parts := strings.Split(uri, "/")
	return parts[len(parts)-1]
}
