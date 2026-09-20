package peer

import (
	"context"

	"bwawan.com/openuss/internal/util"
	"bwawan.com/openuss/internal/utmclient"
	"bwawan.com/openuss/sdk/api/scdussv1"
)

type UTMClient struct {
	Client *utmclient.Client
}

func New(client *utmclient.Client) *UTMClient {
	return &UTMClient{Client: client}
}

func (p *UTMClient) GetOperationalIntentDetails(
	ctx context.Context,
	ussBaseURL scdussv1.OperationalIntentUssBaseURL,
	entityID scdussv1.EntityID,
) (scdussv1.GetOperationalIntentDetailsResponse, error) {
	endpoint := string(ussBaseURL) + "/uss/v1/operational_intents/" + string(entityID)
	response, err := p.Client.Get(ctx, endpoint, scdussv1.UtmStrategicCoordinationScope)
	if err != nil {
		return scdussv1.GetOperationalIntentDetailsResponse{}, err
	}
	return util.UnmarshalType[scdussv1.GetOperationalIntentDetailsResponse](response.Body)
}
