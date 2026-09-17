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

type AirspaceConflictError struct {
	Message                   *string
	MissingOperationalIntents *[]scdussv1.OperationalIntentReference
}

func (err AirspaceConflictError) Error() string {
	return *err.Message
}

func NewAirspaceConflictError(response scdussv1.AirspaceConflictResponse) AirspaceConflictError {
	return AirspaceConflictError{
		Message:                   response.Message,
		MissingOperationalIntents: response.MissingOperationalIntents,
	}
}

func (dss *DSS) PutOperationalIntentReference(
	ctx context.Context,
	entityId scdussv1.EntityID,
	ovn *scdussv1.EntityOVN,
	reference scdussv1.PutOperationalIntentReferenceParameters,
) (scdussv1.ChangeOperationalIntentReferenceResponse, error) {
	endpoint := dss.toOperationalIntentEndpoint(entityId, ovn)
	response, err := dss.Client.Put(ctx, endpoint, reference, scdussv1.UtmStrategicCoordinationScope)
	if err != nil {
		return scdussv1.ChangeOperationalIntentReferenceResponse{}, err
	}
	if response.StatusCode != http.StatusConflict {
		return util.UnmarshalType[scdussv1.ChangeOperationalIntentReferenceResponse](response.Body)
	}

	conflict, _ := util.UnmarshalType[scdussv1.AirspaceConflictResponse](response.Body)
	return scdussv1.ChangeOperationalIntentReferenceResponse{}, NewAirspaceConflictError(conflict)
}

func (dss *DSS) DeleteOperationalIntent(
	ctx context.Context,
	entityId scdussv1.EntityID,
	ovn scdussv1.EntityOVN,
) (scdussv1.ChangeOperationalIntentReferenceResponse, error) {
	endpoint := dss.toOperationalIntentEndpoint(entityId, &ovn)
	response, err := dss.Client.Delete(ctx, endpoint, scdussv1.UtmStrategicCoordinationScope)
	if err != nil {
		return scdussv1.ChangeOperationalIntentReferenceResponse{}, err
	}
	return util.UnmarshalType[scdussv1.ChangeOperationalIntentReferenceResponse](response.Body)
}

func (dss *DSS) toOperationalIntentEndpoint(entityId scdussv1.EntityID, ovn *scdussv1.EntityOVN) string {
	endpoint := dss.Host + "/dss/v1/operational_intent_references/" + string(entityId)
	if ovn != nil {
		endpoint += "/" + string(*ovn)
	}
	return endpoint
}
