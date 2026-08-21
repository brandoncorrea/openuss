package dss

import (
	"context"

	"bwawan.com/openuss/internal/api/scdussv1"
)

type USSAuthority interface {
	CreateOperationalIntentReference(context.Context, scdussv1.EntityID, scdussv1.PutOperationalIntentReferenceParameters) (scdussv1.ChangeOperationalIntentReferenceResponse, error)
}
