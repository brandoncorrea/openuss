package peertest

import (
	"context"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/peer"
)

type Notification struct {
	USSBaseURL scdussv1.SubscriptionUssBaseURL
	Details    scdussv1.PutOperationalIntentDetailsParameters
}

type NotificationRecorder struct {
	peer.Client
	Notifications []Notification
}

func (r *NotificationRecorder) NotifyOperationalIntentDetails(
	_ context.Context,
	ussBaseURL scdussv1.SubscriptionUssBaseURL,
	details scdussv1.PutOperationalIntentDetailsParameters,
) error {
	notification := Notification{
		USSBaseURL: ussBaseURL,
		Details:    details,
	}
	r.Notifications = append(r.Notifications, notification)
	return nil
}
