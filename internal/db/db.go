package db

import (
	"uuid"

	"bwawan.com/openuss/sdk/api/scdussv1"
)

type FlightStore interface {
	Upsert(FlightPlan) error
	Get(uuid.UUID) *FlightPlan
	Delete(uuid.UUID)
	List() []FlightPlan
}

type FlightPlan struct {
	ID       uuid.UUID
	EntityID scdussv1.EntityID
}
