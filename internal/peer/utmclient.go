package peer

import (
	"context"

	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/util"
	"bwawan.com/openuss/internal/utmclient"
)

type UTMClient struct {
	Client *utmclient.Client
}

func New(client *utmclient.Client) *UTMClient {
	return &UTMClient{Client: client}
}

func (peer *UTMClient) GetOperationalIntentDetails(
	ctx context.Context,
	ussBaseUrl scdussv1.OperationalIntentUssBaseURL,
	entityId scdussv1.EntityID,
) (scdussv1.GetOperationalIntentDetailsResponse, error) {
	endpoint := string(ussBaseUrl) + "/uss/v1/operational_intents/" + string(entityId)
	response, err := peer.Client.Get(ctx, endpoint, scdussv1.UtmStrategicCoordinationScope)
	if err != nil {
		return scdussv1.GetOperationalIntentDetailsResponse{}, err
	}
	return util.UnmarshalType[scdussv1.GetOperationalIntentDetailsResponse](response.Body)
}
