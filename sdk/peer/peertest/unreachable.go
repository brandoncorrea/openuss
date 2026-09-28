package peertest

import (
	"context"
	"errors"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/peer"
)

var ErrUnreachable = errors.New("peer unreachable")

type UnreachablePeer struct {
	peer.Client
}

func (UnreachablePeer) GetOperationalIntentDetails(
	context.Context,
	scdussv1.OperationalIntentUssBaseURL,
	scdussv1.EntityID,
) (scdussv1.GetOperationalIntentDetailsResponse, error) {
	return scdussv1.GetOperationalIntentDetailsResponse{}, ErrUnreachable
}
