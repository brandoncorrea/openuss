package db

import (
	"iter"
	"time"
	"uuid"

	"bwawan.com/openuss/internal/api/scdussv1"
)

type DB interface {
	SaveIntent(OperationalIntent) error
	GetIntent(scdussv1.EntityID) *OperationalIntent
	DeleteIntent(scdussv1.EntityID)
	SaveFlight(FlightPlan) error
	GetFlight(uuid.UUID) *FlightPlan
	DeleteFlight(uuid.UUID)
	GetAllFlights() iter.Seq[FlightPlan]
	GetAllIntents() iter.Seq[OperationalIntent]
}

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

type FlightPlan struct {
	ID       uuid.UUID
	EntityID scdussv1.EntityID
}
