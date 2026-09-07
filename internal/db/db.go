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
	UssAvailability scdussv1.UssAvailabilityState
	Version         int32
	State           scdussv1.OperationalIntentState
	Ovn             scdussv1.EntityOVN
	TimeStart       time.Time
	TimeEnd         time.Time
	UssBaseUrl      scdussv1.OperationalIntentUssBaseURL
	SubscriptionId  scdussv1.SubscriptionID
	Volumes         []scdussv1.Volume4D
}

type FlightPlan struct {
	Id       uuid.UUID
	EntityID scdussv1.EntityID
}
