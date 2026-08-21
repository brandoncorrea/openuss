package dss

import (
	"context"
	"fmt"
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
	response, err := dss.MakeRequest(ctx, http.MethodPut, uri, reference, scdussv1.UtmStrategicCoordinationScope)
	if err != nil {
		return scdussv1.ChangeOperationalIntentReferenceResponse{}, err
	}
	result, err := util.UnmarshalType[scdussv1.ChangeOperationalIntentReferenceResponse](response.Body)
	if err != nil {
		return result, fmt.Errorf("dss: failed to parse response body: %w", err)
	}
	return result, nil
}
