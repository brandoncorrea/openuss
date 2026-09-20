package dss

import (
	"context"
	"net/http"

	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/util"
	"bwawan.com/openuss/internal/utmclient"
)

type DSS struct {
	Host   string
	Client *utmclient.Client
}

func New(host string, client *utmclient.Client) *DSS {
	return &DSS{
		Host:   host,
		Client: client,
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

func (d *DSS) PutOperationalIntentReference(
	ctx context.Context,
	entityID scdussv1.EntityID,
	ovn *scdussv1.EntityOVN,
	reference scdussv1.PutOperationalIntentReferenceParameters,
) (scdussv1.ChangeOperationalIntentReferenceResponse, error) {
	endpoint := d.toOperationalIntentEndpoint(entityID, ovn)
	response, err := d.Client.Put(ctx, endpoint, reference, scdussv1.UtmStrategicCoordinationScope)
	if err != nil {
		return scdussv1.ChangeOperationalIntentReferenceResponse{}, err
	}
	if response.StatusCode != http.StatusConflict {
		return util.UnmarshalType[scdussv1.ChangeOperationalIntentReferenceResponse](response.Body)
	}

	conflict, _ := util.UnmarshalType[scdussv1.AirspaceConflictResponse](response.Body)
	return scdussv1.ChangeOperationalIntentReferenceResponse{}, NewAirspaceConflictError(conflict)
}

func (d *DSS) DeleteOperationalIntentReference(
	ctx context.Context,
	entityID scdussv1.EntityID,
	ovn scdussv1.EntityOVN,
) (scdussv1.ChangeOperationalIntentReferenceResponse, error) {
	endpoint := d.toOperationalIntentEndpoint(entityID, &ovn)
	response, err := d.Client.Delete(ctx, endpoint, scdussv1.UtmStrategicCoordinationScope)
	if err != nil {
		return scdussv1.ChangeOperationalIntentReferenceResponse{}, err
	}
	return util.UnmarshalType[scdussv1.ChangeOperationalIntentReferenceResponse](response.Body)
}

func (d *DSS) toOperationalIntentEndpoint(entityID scdussv1.EntityID, ovn *scdussv1.EntityOVN) string {
	endpoint := d.Host + "/dss/v1/operational_intent_references/" + string(entityID)
	if ovn != nil {
		endpoint += "/" + string(*ovn)
	}
	return endpoint
}
