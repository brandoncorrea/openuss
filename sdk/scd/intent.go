package scd

import (
	"time"

	"bwawan.com/openuss/sdk/api/scdussv1"
)

type OperationalIntent struct {
	EntityID        scdussv1.EntityID
	Manager         string
	USSAvailability scdussv1.UssAvailabilityState
	Version         int32
	Priority        scdussv1.Priority
	State           scdussv1.OperationalIntentState
	OVN             scdussv1.EntityOVN
	TimeStart       time.Time
	TimeEnd         time.Time
	USSBaseURL      scdussv1.OperationalIntentUssBaseURL
	SubscriptionID  scdussv1.SubscriptionID
	Volumes         []scdussv1.Volume4D
}
