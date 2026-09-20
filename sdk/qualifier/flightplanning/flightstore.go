package flightplanning

import (
	"uuid"

	"bwawan.com/openuss/sdk/api/scdussv1"
)

type FlightStore interface {
	Upsert(FlightPlanRecord) error
	Get(uuid.UUID) *FlightPlanRecord
	Delete(uuid.UUID)
	List() []FlightPlanRecord
}

type FlightPlanRecord struct {
	ID       uuid.UUID
	EntityID scdussv1.EntityID
}
