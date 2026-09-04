package dss

import (
	"context"
	"net/http"

	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/auth"
	"bwawan.com/openuss/internal/util"
)

type DSS struct {
	Host        string
	Audience    string
	TokenSource auth.TokenSource
	Client      *http.Client
}

func (dss *DSS) CreateOperationalIntentReference(
	ctx context.Context,
	entityId scdussv1.EntityID,
	reference scdussv1.PutOperationalIntentReferenceParameters,
) (scdussv1.ChangeOperationalIntentReferenceResponse, error) {
	uri := "/dss/v1/operational_intent_references/" + string(entityId)
	return requestIntentChange(dss, ctx, http.MethodPut, uri, reference)
}

func (dss *DSS) DeleteOperationalIntent(
	ctx context.Context,
	entityId scdussv1.EntityID,
	ovn scdussv1.EntityOVN,
) (scdussv1.ChangeOperationalIntentReferenceResponse, error) {
	uri := "/dss/v1/operational_intent_references/" + string(entityId) + "/" + string(ovn)
	return requestIntentChange(dss, ctx, http.MethodDelete, uri, nil)
}

func requestIntentChange(
	dss *DSS,
	ctx context.Context,
	method string,
	uri string,
	body any,
) (scdussv1.ChangeOperationalIntentReferenceResponse, error) {
	response, err := dss.MakeRequest(ctx, method, uri, body, scdussv1.UtmStrategicCoordinationScope)
	if err != nil {
		return scdussv1.ChangeOperationalIntentReferenceResponse{}, err
	}
	return util.UnmarshalType[scdussv1.ChangeOperationalIntentReferenceResponse](response.Body)
}
