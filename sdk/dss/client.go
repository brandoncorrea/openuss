package dss

import (
	"context"

	"bwawan.com/openuss/sdk/api/scdussv1"
)

type Client interface {
	PutOperationalIntentReference(
		context.Context,
		scdussv1.EntityID,
		*scdussv1.EntityOVN,
		scdussv1.PutOperationalIntentReferenceParameters,
	) (scdussv1.ChangeOperationalIntentReferenceResponse, error)

	DeleteOperationalIntentReference(
		context.Context,
		scdussv1.EntityID,
		scdussv1.EntityOVN,
	) (scdussv1.ChangeOperationalIntentReferenceResponse, error)
}
