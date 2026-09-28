package scd

import (
	"context"
	"slices"

	"bwawan.com/openuss/sdk/api/scdussv1"
)

// TODO(gap): Only the first peer subscriber is notified - the rest are ignored
func (s *Service) notifyPeers(
	ctx context.Context,
	result scdussv1.ChangeOperationalIntentReferenceResponse,
) {
	ownURL := scdussv1.SubscriptionUssBaseURL(s.USSBaseURL)
	index := slices.IndexFunc(result.Subscribers, func(subscriber scdussv1.SubscriberToNotify) bool {
		return subscriber.UssBaseUrl != ownURL
	})

	if index >= 0 {
		s.notifyPeer(ctx, result.Subscribers[index], result)
	}
}

func (s *Service) notifyPeer(
	ctx context.Context,
	subscriber scdussv1.SubscriberToNotify,
	result scdussv1.ChangeOperationalIntentReferenceResponse,
) {
	// TODO(gap): No error handling
	s.Peer.NotifyOperationalIntentDetails(
		ctx,
		subscriber.UssBaseUrl,
		// TODO(gap): Missing 'Subscriptions' attribute
		scdussv1.PutOperationalIntentDetailsParameters{
			OperationalIntentId: result.OperationalIntentReference.Id,
			// TODO(gap): Missing 'Details' attribute
			OperationalIntent: &scdussv1.OperationalIntent{
				Reference: scdussv1.OperationalIntentReference{
					// TODO(gap): Missing 'Version', 'Ovn', 'UssBaseUrl', and 'SubscriptionId' attributes
					Id:              result.OperationalIntentReference.Id,
					Manager:         result.OperationalIntentReference.Manager,
					UssAvailability: result.OperationalIntentReference.UssAvailability,
					State:           result.OperationalIntentReference.State,
					TimeStart:       result.OperationalIntentReference.TimeStart,
					TimeEnd:         result.OperationalIntentReference.TimeEnd,
				},
			},
		})
}
