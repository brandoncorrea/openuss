package dsstest

import (
	"encoding/json/v2"
	"net/http"
	"slices"
	"strings"
	"testing"
	"uuid"

	"bwawan.com/openuss/sdk/api"
	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/dss"
	"bwawan.com/openuss/sdk/internal/util"
	"bwawan.com/openuss/sdk/utmclient/utmclienttest"
)

const (
	Host    = "dss.localutm"
	BaseURL = "http://" + Host
)

func NewDSS(t *testing.T, handler http.HandlerFunc) *dss.UTMClient {
	t.Helper()
	return dss.New(BaseURL, utmclienttest.NewClient(t, handler))
}

func NewEcosystemHandler(registered []scdussv1.OperationalIntent) http.HandlerFunc {
	handleDSS := dssHandlerFor(registered)
	handleUSS := ussHandlerFor(registered)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Host == Host {
			handleDSS(w, r)
		} else {
			handleUSS(w, r)
		}
	}
}

func ussHandlerFor(registered []scdussv1.OperationalIntent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		index := slices.IndexFunc(registered, func(intent scdussv1.OperationalIntent) bool {
			host := strings.TrimPrefix(string(intent.Reference.UssBaseUrl), "http://")
			return host == r.Host && string(intent.Reference.Id) == ussEntityID(r.RequestURI)
		})

		if index >= 0 {
			api.WriteJSON(w, http.StatusOK, scdussv1.GetOperationalIntentDetailsResponse{
				OperationalIntent: registered[index],
			})
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func dssHandlerFor(registered []scdussv1.OperationalIntent) http.HandlerFunc {
	references := util.Map(registered, func(intent scdussv1.OperationalIntent) scdussv1.OperationalIntentReference {
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
					Id:         scdussv1.EntityID(dssEntityID(r.RequestURI)),
					Ovn:        new(scdussv1.EntityOVN(uuid.New().String())),
					UssBaseUrl: putParams.UssBaseUrl,
				},
			})
		}
	}
}

const ussIntentsPath = "/uss/v1/operational_intents/"

func ussEntityID(uri string) string {
	return strings.TrimPrefix(uri, ussIntentsPath)
}

const dssIntentReferencesPath = "/dss/v1/operational_intent_references/"

func dssEntityID(uri string) string {
	entityID, _, _ := strings.Cut(strings.TrimPrefix(uri, dssIntentReferencesPath), "/")
	return entityID
}
