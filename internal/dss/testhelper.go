package dss

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"bwawan.com/openuss/internal/api"
	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/auth"
	"bwawan.com/openuss/internal/peer"
	"bwawan.com/openuss/internal/util"
	"bwawan.com/openuss/internal/utmclient"
)

var fakeDssHost = "http://dss.example.com"

func newDss(t *testing.T, handler http.HandlerFunc) *DSS {
	server := httptest.NewTestServer(t, http.HandlerFunc(handler))
	return &DSS{
		Host:   fakeDssHost,
		Client: utmclient.New(auth.NewInMemoryTokenSource(), server.Client()),
		Peer:   peer.NewInMemoryPeer([]scdussv1.OperationalIntentReference{}),
	}
}

func newConflictingDss(t *testing.T, peers []scdussv1.OperationalIntentReference) *DSS {
	peer := peer.NewInMemoryPeer(peers)
	dss := newDss(t, newConflictHandler(peer))
	dss.Peer = peer
	return dss
}

func newConflictHandler(peer *peer.InMemoryPeer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(fakeDssHost, r.Host) {
			handleDssConflict(w, r, peer)
		} else {
			handleGetUssIntent(w, r, peer)
		}
	}
}

func handleDssConflict(w http.ResponseWriter, r *http.Request, peer *peer.InMemoryPeer) {
	params, err := util.UnmarshalReadType[scdussv1.PutOperationalIntentReferenceParameters](r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if hasAllOvns(params.Key, peer.References) {
		api.WriteJSON(w, http.StatusOK, scdussv1.ChangeOperationalIntentReferenceResponse{
			OperationalIntentReference: scdussv1.OperationalIntentReference{
				Id: parseEntityId(r),
			},
		})
	} else {
		api.WriteJSON(w, http.StatusConflict, scdussv1.AirspaceConflictResponse{
			MissingOperationalIntents: new(slices.Clone(peer.References)),
		})
	}
}

func handleGetUssIntent(w http.ResponseWriter, r *http.Request, peer *peer.InMemoryPeer) {
	baseUrl := scdussv1.OperationalIntentUssBaseURL("http://" + r.Host)
	result, err := peer.GetOperationalIntentDetails(r.Context(), baseUrl, parseEntityId(r))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		api.WriteJSON(w, http.StatusOK, result)
	}
}

func hasAllOvns(key *scdussv1.Key, peers []scdussv1.OperationalIntentReference) bool {
	ovns := util.Map(peers, getIntentOvn)
	return key != nil && slices.Equal(*key, ovns)
}

func getIntentOvn(intent scdussv1.OperationalIntentReference) scdussv1.EntityOVN {
	return *intent.Ovn
}

func parseEntityId(r *http.Request) scdussv1.EntityID {
	uriParts := strings.Split(r.RequestURI, "/")
	return scdussv1.EntityID(uriParts[len(uriParts)-1])
}
