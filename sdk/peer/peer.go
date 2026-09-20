package peer

import (
	"context"

	"bwawan.com/openuss/sdk/api/scdussv1"
)

type Client interface {
	GetOperationalIntentDetails(
		context.Context,
		scdussv1.OperationalIntentUssBaseURL,
		scdussv1.EntityID,
	) (scdussv1.GetOperationalIntentDetailsResponse, error)
}
