package db

import (
	"uuid"

	"bwawan.com/openuss/sdk/api/scdussv1"
)

type DB interface {
	SaveFlight(FlightPlan) error
	GetFlight(uuid.UUID) *FlightPlan
	DeleteFlight(uuid.UUID)
	GetAllFlights() []FlightPlan
}

type FlightPlan struct {
	ID       uuid.UUID
	EntityID scdussv1.EntityID
}
