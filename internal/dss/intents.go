package dss

import (
	"context"
	"net/http"

	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/peer"
	"bwawan.com/openuss/internal/util"
	"bwawan.com/openuss/internal/utmclient"
)

type DSS struct {
	Host   string
	Client *utmclient.Client
	Peer   peer.Client
}

func (dss *DSS) CreateOperationalIntentReference(
	ctx context.Context,
	entityId scdussv1.EntityID,
	reference scdussv1.PutOperationalIntentReferenceParameters,
) (scdussv1.ChangeOperationalIntentReferenceResponse, error) {
	uri := "/dss/v1/operational_intent_references/" + string(entityId)
	response, err := dss.Client.Put(ctx, dss.Host+uri, reference, scdussv1.UtmStrategicCoordinationScope)
	if err != nil {
		return scdussv1.ChangeOperationalIntentReferenceResponse{}, err
	}
	if response.StatusCode != http.StatusConflict {
		return util.UnmarshalType[scdussv1.ChangeOperationalIntentReferenceResponse](response.Body)
	}

	conflict, _ := util.UnmarshalType[scdussv1.AirspaceConflictResponse](response.Body)
	intent := (*conflict.MissingOperationalIntents)[0]
	details, _ := dss.Peer.GetOperationalIntentDetails(ctx, intent.UssBaseUrl, intent.Id)
	reference.Key = &scdussv1.Key{*details.OperationalIntent.Reference.Ovn}
	response, _ = dss.Client.Put(ctx, dss.Host+uri, reference, scdussv1.UtmStrategicCoordinationScope)
	return util.UnmarshalType[scdussv1.ChangeOperationalIntentReferenceResponse](response.Body)
}

func (dss *DSS) DeleteOperationalIntent(
	ctx context.Context,
	entityId scdussv1.EntityID,
	ovn scdussv1.EntityOVN,
) (scdussv1.ChangeOperationalIntentReferenceResponse, error) {
	uri := "/dss/v1/operational_intent_references/" + string(entityId) + "/" + string(ovn)
	response, err := dss.Client.Delete(ctx, dss.Host+uri, scdussv1.UtmStrategicCoordinationScope)
	if err != nil {
		return scdussv1.ChangeOperationalIntentReferenceResponse{}, err
	}
	return util.UnmarshalType[scdussv1.ChangeOperationalIntentReferenceResponse](response.Body)
}
