package flightplanning

import (
	"context"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/scd"
)

type StrategicCoordination interface {
	CreateOperationalIntent(context.Context, scd.IntentParams) (scd.OperationalIntent, error)
	UpdateOperationalIntent(
		context.Context,
		scdussv1.EntityID,
		scd.IntentParams,
	) (scd.OperationalIntent, error)
	DeleteOperationalIntent(context.Context, scdussv1.EntityID) error
}

type Handler struct {
	SCD     StrategicCoordination
	Flights FlightStore
}

func New(service StrategicCoordination, flights FlightStore) *Handler {
	return &Handler{
		SCD:     service,
		Flights: flights,
	}
}
