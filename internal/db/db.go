package db

import (
	"iter"
	"uuid"

	"bwawan.com/openuss/internal/api/scdussv1"
)

type DB interface {
	SaveFlight(FlightPlan) error
	GetFlight(uuid.UUID) *FlightPlan
	DeleteFlight(uuid.UUID)
	GetAllFlights() iter.Seq[FlightPlan]
}

type FlightPlan struct {
	ID       uuid.UUID
	EntityID scdussv1.EntityID
}
