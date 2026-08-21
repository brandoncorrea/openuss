package dss

import (
	"context"

	"bwawan.com/openuss/internal/api/scdussv1"
)

type InMemoryDSS struct {
	References map[scdussv1.EntityID]scdussv1.PutOperationalIntentReferenceParameters
}

func NewInMemoryDSS() *InMemoryDSS {
	return &InMemoryDSS{
		References: map[scdussv1.EntityID]scdussv1.PutOperationalIntentReferenceParameters{},
	}
}

func (dss *InMemoryDSS) CreateOperationalIntentReference(
	ctx context.Context,
	id scdussv1.EntityID,
	reference scdussv1.PutOperationalIntentReferenceParameters,
) (scdussv1.ChangeOperationalIntentReferenceResponse, error) {
	if ctx == nil {
		panic("inmemory dss: missing Context")
	}
	dss.References[id] = reference
	return scdussv1.ChangeOperationalIntentReferenceResponse{}, nil
}
