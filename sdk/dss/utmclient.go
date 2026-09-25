package dss

import (
	"context"
	"net/http"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/internal/util"
	"bwawan.com/openuss/sdk/utmclient"
)

type UTMClient struct {
	BaseURL string
	Client  *utmclient.Client
}

func New(baseURL string, client *utmclient.Client) *UTMClient {
	return &UTMClient{
		BaseURL: baseURL,
		Client:  client,
	}
}

type AirspaceConflictError struct {
	Message                   *string
	MissingOperationalIntents *[]scdussv1.OperationalIntentReference
}

func (e AirspaceConflictError) Error() string {
	return *e.Message
}

func NewAirspaceConflictError(response scdussv1.AirspaceConflictResponse) AirspaceConflictError {
	return AirspaceConflictError{
		Message:                   response.Message,
		MissingOperationalIntents: response.MissingOperationalIntents,
	}
}

func (u *UTMClient) PutOperationalIntentReference(
	ctx context.Context,
	entityID scdussv1.EntityID,
	ovn *scdussv1.EntityOVN,
	reference scdussv1.PutOperationalIntentReferenceParameters,
) (scdussv1.ChangeOperationalIntentReferenceResponse, error) {
	endpoint := u.toOperationalIntentEndpoint(entityID, ovn)
	response, err := u.Client.Put(ctx, endpoint, reference, scdussv1.UtmStrategicCoordinationScope, scdussv1.UtmConformanceMonitoringSaScope)
	if err != nil {
		return scdussv1.ChangeOperationalIntentReferenceResponse{}, err
	}
	if response.StatusCode != http.StatusConflict {
		return util.UnmarshalType[scdussv1.ChangeOperationalIntentReferenceResponse](response.Body)
	}

	conflict, _ := util.UnmarshalType[scdussv1.AirspaceConflictResponse](response.Body)
	return scdussv1.ChangeOperationalIntentReferenceResponse{}, NewAirspaceConflictError(conflict)
}

func (u *UTMClient) DeleteOperationalIntentReference(
	ctx context.Context,
	entityID scdussv1.EntityID,
	ovn scdussv1.EntityOVN,
) (scdussv1.ChangeOperationalIntentReferenceResponse, error) {
	endpoint := u.toOperationalIntentEndpoint(entityID, &ovn)
	response, err := u.Client.Delete(ctx, endpoint, scdussv1.UtmStrategicCoordinationScope)
	if err != nil {
		return scdussv1.ChangeOperationalIntentReferenceResponse{}, err
	}
	return util.UnmarshalType[scdussv1.ChangeOperationalIntentReferenceResponse](response.Body)
}

func (u *UTMClient) toOperationalIntentEndpoint(entityID scdussv1.EntityID, ovn *scdussv1.EntityOVN) string {
	endpoint := u.BaseURL + "/dss/v1/operational_intent_references/" + string(entityID)
	if ovn != nil {
		endpoint += "/" + string(*ovn)
	}
	return endpoint
}
